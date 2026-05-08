'use client';

import React, { useState, useEffect, useRef } from 'react';
import { useMarketData } from '@/hooks/useMarketData';
import PriceChart from './PriceChart';
import { Activity, TrendingUp, TrendingDown, Gauge, Clock } from 'lucide-react';
import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';
import { Time } from 'lightweight-charts';

function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

const Dashboard = () => {
  const { data, status } = useMarketData('ws://localhost:8080/ws');
  const [history, setHistory] = useState<{ time: Time; value: number }[]>([]);
  const [prevPrice, setPrevPrice] = useState<number>(0);
  const [priceColor, setPriceColor] = useState<'neutral' | 'up' | 'down'>('neutral');
  const timeoutRef = useRef<NodeJS.Timeout | null>(null);

  // Fetch initial history from TimescaleDB
  useEffect(() => {
    const fetchHistory = async () => {
      try {
        const res = await fetch('http://localhost:8080/api/history');
        const initialData = await res.json();
        if (Array.isArray(initialData)) {
          let lastTime = 0;
          const formatted = initialData.reverse().map((d: any) => {
            let t = Math.floor(new Date(d.updated_at).getTime() / 1000);
            if (t <= lastTime) t = lastTime + 1;
            lastTime = t;
            return {
              time: t as Time,
              value: parseFloat(d.price)
            };
          });
          setHistory(formatted);
          if (formatted.length > 0) {
            setPrevPrice(formatted[formatted.length - 1].value);
          }
        }
      } catch (err) {
        console.error("Failed to fetch history:", err);
      }
    };
    fetchHistory();
  }, []);

  useEffect(() => {
    if (data) {
      const currentPrice = parseFloat(data.price);

      if (currentPrice > prevPrice) {
        setPriceColor('up');
      } else if (currentPrice < prevPrice) {
        setPriceColor('down');
      }

      setPrevPrice(currentPrice);

      if (timeoutRef.current) clearTimeout(timeoutRef.current);
      timeoutRef.current = setTimeout(() => setPriceColor('neutral'), 300);

      const time = Math.floor(new Date(data.updated_at).getTime() / 1000) as Time;
      setHistory(prev => {
        let time = Math.floor(new Date(data.updated_at).getTime() / 1000) as Time;

        if (prev.length > 0 && (time as number) <= (prev[prev.length - 1].time as number)) {
          time = ((prev[prev.length - 1].time as number) + 1) as Time;
        }

        const newHistory = [...prev, { time, value: currentPrice }];
        return newHistory.slice(-200);
      });
    }
  }, [data, prevPrice]);

  const sentiment = data?.sentiment ?? 0;
  const sentimentPos = ((sentiment + 1) / 2) * 100;

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 p-4 md:p-8 font-sans selection:bg-blue-500/30">
      <div className="max-w-6xl mx-auto space-y-8">
        <header className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-white/5 pb-8">
          <div>
            <h1 className="text-4xl font-black tracking-tighter bg-gradient-to-br from-white via-white to-slate-500 bg-clip-text text-transparent">
              SENTIGRAPH<span className="text-blue-500">.</span>ANALYST
            </h1>
            <p className="text-slate-500 mt-1 flex items-center gap-2 font-medium">
              <span className={cn(
                "w-2 h-2 rounded-full",
                status === 'open' ? "bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.5)] animate-pulse" : "bg-rose-500"
              )} />
              {status === 'open' ? 'Live BTC/USDT Feed Active' : 'Connecting to high-speed stream...'}
            </p>
          </div>
          <div className="flex items-center gap-3">
            <div className="bg-slate-900/80 backdrop-blur-md px-4 py-2 rounded-2xl border border-white/5 flex items-center gap-2">
              <Clock size={14} className="text-slate-500" />
              <span className="text-[10px] font-bold text-slate-400 tracking-widest uppercase">
                {new Date().toLocaleTimeString()}
              </span>
            </div>
            <div className="bg-blue-500/10 text-blue-400 px-4 py-2 rounded-2xl border border-blue-500/20 text-[10px] font-bold tracking-widest uppercase">
              Pro Terminal
            </div>
          </div>
        </header>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="bg-slate-900/40 backdrop-blur-2xl border border-white/10 rounded-[2rem] p-8 shadow-2xl relative overflow-hidden group">
            <div className="absolute -top-24 -right-24 w-64 h-64 bg-blue-500/10 rounded-full blur-[80px]" />
            <div className="flex items-center gap-2 text-slate-400 font-bold text-[10px] tracking-widest uppercase mb-4">
              <Activity size={14} className="text-blue-400" />
              Spot Price Index
            </div>
            <div className={cn(
              "text-6xl md:text-7xl font-mono font-bold tracking-tighter transition-all duration-300",
              priceColor === 'up' ? "text-emerald-400 scale-[1.02]" :
                priceColor === 'down' ? "text-rose-400 scale-[0.98]" : "text-white"
            )}>
              <span className="opacity-30 text-4xl mr-2">$</span>
              {data?.price ? parseFloat(data.price).toLocaleString(undefined, { minimumFractionDigits: 2 }) : '---.--'}
            </div>
          </div>

          <div className="bg-slate-900/40 backdrop-blur-2xl border border-white/10 rounded-[2rem] p-8 shadow-2xl relative overflow-hidden group">
            <div className="absolute -top-24 -right-24 w-64 h-64 bg-purple-500/10 rounded-full blur-[80px]" />
            <div className="flex items-center gap-2 text-slate-400 font-bold text-[10px] tracking-widest uppercase mb-4">
              <Gauge size={14} className="text-purple-400" />
              AI Market Sentiment
            </div>
            <div className="text-6xl md:text-7xl font-mono font-bold text-white tracking-tighter mb-8">
              {sentiment > 0 ? '+' : ''}{sentiment.toFixed(2)}
            </div>
            <div className="space-y-3">
              <div className="h-3 w-full bg-slate-800/50 rounded-full overflow-hidden p-0.5 border border-white/5">
                <div
                  className="h-full bg-gradient-to-r from-rose-500 via-slate-400 to-emerald-500 rounded-full transition-all duration-1000 ease-in-out shadow-[0_0_15px_rgba(59,130,246,0.3)]"
                  style={{ width: `${sentimentPos}%` }}
                />
              </div>
              <div className="flex justify-between text-[9px] font-black text-slate-500 tracking-widest uppercase">
                <span className={sentiment < -0.3 ? "text-rose-400" : ""}>Bearish</span>
                <span className={sentiment > -0.3 && sentiment < 0.3 ? "text-white" : ""}>Neutral</span>
                <span className={sentiment > 0.3 ? "text-emerald-400" : ""}>Bullish</span>
              </div>
            </div>
          </div>
        </div>

        <div className="bg-slate-900/40 backdrop-blur-2xl border border-white/10 rounded-[2.5rem] p-1 shadow-2xl">
          <div className="p-8">
            <div className="flex items-center justify-between mb-8">
              <div className="flex items-center gap-2 text-slate-400 font-bold text-[10px] tracking-widest uppercase">
                <TrendingUp size={14} className="text-emerald-400" />
                Live Performance Chart
              </div>
              <div className="flex gap-2">
                <div className="h-2 w-2 rounded-full bg-emerald-500 animate-ping" />
                <span className="text-[10px] font-bold text-emerald-500 uppercase tracking-widest">Real-time</span>
              </div>
            </div>
            <PriceChart data={history} />
          </div>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
