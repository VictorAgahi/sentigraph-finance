'use client';

import { useState, useEffect } from 'react';
import { useMarketData, LiquidationData, MarketData } from '@/hooks/useMarketData';
import CandleChart from './CandleChart';
import { Activity, TrendingUp, Gauge, Database, RefreshCw, Trash2 } from 'lucide-react';
import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';
import { CandlestickData, Time } from 'lightweight-charts';

function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

type Asset = 'btc' | 'eth' | 'sol';

const Dashboard = () => {
  const [selectedAsset, setSelectedAsset] = useState<Asset>('btc');
  const { data, status } = useMarketData('ws://localhost:8080/ws');

  // Store candlestick history for each asset
  const [assetCandles, setAssetCandles] = useState<Record<Asset, CandlestickData<Time>[]>>({
    btc: [], eth: [], sol: []
  });

  const [timeframe, setTimeframe] = useState<number>(60); // Default 1m in seconds

  const [currentPrices, setCurrentPrices] = useState<Record<Asset, number>>({
    btc: 0, eth: 0, sol: 0
  });

  const [sentiments, setSentiments] = useState<Record<Asset, number>>({
    btc: 0, eth: 0, sol: 0
  });

  const [imbalances, setImbalances] = useState<Record<Asset, number>>({
    btc: 1, eth: 1, sol: 1
  });

  const [liquidations, setLiquidations] = useState<LiquidationData[]>([]);

  useEffect(() => {
    // Initial load from localStorage
    const savedCandles = localStorage.getItem('sentigraph_candles');
    const savedLiqs = localStorage.getItem('sentigraph_liquidations');
    if (savedCandles) {
      try { setAssetCandles(JSON.parse(savedCandles)); } catch (e) {}
    }
    if (savedLiqs) {
      try { setLiquidations(JSON.parse(savedLiqs)); } catch (e) {}
    }
  }, []);

  useEffect(() => {
    const fetchHistory = async () => {
      try {
        const response = await fetch(`http://localhost:8080/api/history/${selectedAsset}`);
        if (!response.ok) throw new Error('History fetch failed');
        const history: MarketData[] = await response.json();

        if (history && history.length > 0) {
          const reconstructedCandles: CandlestickData<Time>[] = [];
          const candleMap: Record<number, CandlestickData<Time>> = {};

          // History is already ASC from API
          history.forEach((d: MarketData) => {
            const price = parseFloat(d.price);
            const time = (Math.floor(new Date(d.updated_at).getTime() / (timeframe * 1000)) * timeframe) as number;

            if (candleMap[time]) {
              const candle = candleMap[time];
              candle.high = Math.max(candle.high, price);
              candle.low = Math.min(candle.low, price);
              candle.close = price;
            } else {
              candleMap[time] = {
                time: time as Time,
                open: price, high: price, low: price, close: price
              };
              reconstructedCandles.push(candleMap[time]);
            }
          });

          setAssetCandles(prev => ({
            ...prev,
            [selectedAsset]: reconstructedCandles.slice(-200)
          }));

          const latest = history[history.length - 1];
          setCurrentPrices(prev => ({ ...prev, [selectedAsset]: parseFloat(latest.price) }));
          setSentiments(prev => ({ ...prev, [selectedAsset]: latest.sentiment }));
          setImbalances(prev => ({ ...prev, [selectedAsset]: latest.imbalance }));
        }
      } catch (e) {
        console.error("Failed to load history from DB", e);
      }
    };

    fetchHistory();
  }, [selectedAsset]);

  useEffect(() => {
    const timeout = setTimeout(() => {
      localStorage.setItem('sentigraph_candles', JSON.stringify(assetCandles));
      localStorage.setItem('sentigraph_liquidations', JSON.stringify(liquidations));
    }, 1000);
    return () => clearTimeout(timeout);
  }, [assetCandles, liquidations]);

  useEffect(() => {
    if (!data) return;
    console.log("📡 WS INBOUND:", data.type || 'market', data);

    if (data.type === 'liquidation') {
      setLiquidations(prev => [data, ...prev].slice(0, 50));
      return;
    }
    const sym = data.symbol as Asset;
    if (!sym || !data.price) return;

    const price = parseFloat(data.price);
    const time = (Math.floor(Date.now() / (timeframe * 1000)) * timeframe) as Time;

    setCurrentPrices(prev => ({ ...prev, [sym]: price }));
    setSentiments(prev => ({ ...prev, [sym]: data.sentiment }));
    setImbalances(prev => ({ ...prev, [sym]: data.imbalance || 1 }));

    setAssetCandles(prev => {
      const candles = [...prev[sym]];
      const lastCandle = candles[candles.length - 1];

      if (lastCandle && lastCandle.time === time) {
        lastCandle.high = Math.max(lastCandle.high, price);
        lastCandle.low = Math.min(lastCandle.low, price);
        lastCandle.close = price;
        return { ...prev, [sym]: candles };
      } else {
        const newCandle: CandlestickData<Time> = {
          time: time,
          open: price,
          high: price,
          low: price,
          close: price
        };
        return {
          ...prev,
          [sym]: [...candles, newCandle].slice(-200)
        };
      }
    });
  }, [data]);

  const clearHistory = () => {
    const fresh = { btc: [], eth: [], sol: [] };
    setAssetCandles(fresh);
    localStorage.removeItem('sentigraph_candles');
  };

  const currentPrice = currentPrices[selectedAsset];
  const currentSentiment = sentiments[selectedAsset];
  const currentImbalance = imbalances[selectedAsset];
  const candles = assetCandles[selectedAsset];
  const sentimentPos = ((currentSentiment + 1) / 2) * 100;

  const assets: { id: Asset; name: string; color: string }[] = [
    { id: 'btc', name: 'Bitcoin', color: 'text-orange-500' },
    { id: 'eth', name: 'Ethereum', color: 'text-blue-400' },
    { id: 'sol', name: 'Solana', color: 'text-purple-400' },
  ];

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 p-4 md:p-8 font-sans selection:bg-blue-500/30">
      <div className="max-w-6xl mx-auto space-y-8">
        <header className="flex flex-col md:flex-row md:items-center justify-between gap-6 border-b border-white/5 pb-8">
          <div>
            <h1 className="text-4xl font-black tracking-tighter flex items-center gap-2">
              SENTIGRAPH<span className="text-blue-500 text-6xl">.</span>
              <span className="bg-white/5 px-4 py-1 rounded-2xl border border-white/10 text-2xl font-bold ml-2">TERMINAL</span>
            </h1>
            <div className="flex items-center gap-4 mt-3">
              {assets.map(asset => (
                <button
                  key={asset.id}
                  onClick={() => setSelectedAsset(asset.id)}
                  className={cn(
                    "px-4 py-2 rounded-xl text-xs font-black tracking-widest uppercase transition-all border",
                    selectedAsset === asset.id
                      ? "bg-white text-black border-white shadow-[0_0_20px_rgba(255,255,255,0.2)]"
                      : "bg-white/5 text-slate-500 border-white/5 hover:border-white/20"
                  )}
                >
                  {asset.name}
                </button>
              ))}
            </div>
          </div>

          <div className="flex flex-col md:flex-row items-center gap-4">
            <div className="flex items-center gap-4 bg-slate-900/80 backdrop-blur-md px-4 py-2 rounded-2xl border border-white/5">
              <span className="text-[9px] font-black text-slate-500 tracking-widest uppercase">Interval</span>
              <div className="flex gap-1">
                {[1, 5, 15].map(m => (
                  <button
                    key={m}
                    onClick={() => setTimeframe(m * 60)}
                    className={cn(
                      "px-3 py-1 rounded-lg text-[10px] font-bold transition-all",
                      timeframe === m * 60 ? "bg-blue-500 text-white" : "text-slate-500 hover:text-slate-300"
                    )}
                  >
                    {m}M
                  </button>
                ))}
              </div>
            </div>

            <div className="flex items-center gap-3">
              <button
                onClick={clearHistory}
                className="p-3 bg-white/5 rounded-2xl border border-white/5 text-slate-500 hover:text-rose-500 hover:border-rose-500/20 transition-all group"
                title="Clear Cache"
              >
                <Trash2 size={20} className="group-hover:scale-110 transition-transform" />
              </button>
              <div className="bg-slate-900/80 backdrop-blur-md px-4 py-3 rounded-2xl border border-white/5 flex flex-col items-end">
                <span className="text-[9px] font-black text-slate-500 tracking-widest uppercase">Connectivity</span>
                <div className="flex items-center gap-2 mt-1">
                  <span className={cn("w-2 h-2 rounded-full", status === 'open' ? "bg-emerald-500 animate-pulse" : "bg-rose-500")} />
                  <span className="text-xs font-mono font-bold tracking-tighter">{status === 'open' ? 'WS_LIVE' : 'OFFLINE'}</span>
                </div>
              </div>
            </div>
          </div>
        </header>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div className="md:col-span-1 bg-slate-900/40 backdrop-blur-2xl border border-white/10 rounded-[2.5rem] p-8 shadow-2xl relative overflow-hidden">
            <div className="flex items-center gap-2 text-slate-500 font-black text-[10px] tracking-widest uppercase mb-4">
              <Activity size={14} className="text-blue-500" />
              {selectedAsset.toUpperCase()} / USDT
            </div>
            <div className="text-5xl font-mono font-bold tracking-tighter">
              <span className="text-2xl text-slate-600 mr-1">$</span>
              {currentPrice.toLocaleString(undefined, { minimumFractionDigits: 2 })}
            </div>
            <div className="mt-8 pt-8 border-t border-white/5 flex justify-between items-center">
              <div className="space-y-1">
                <span className="text-[9px] font-black text-slate-500 tracking-widest uppercase block">Order Book Imbalance</span>
                <span className={cn(
                  "text-xl font-mono font-bold",
                  currentImbalance > 1.2 ? "text-emerald-400" : currentImbalance < 0.8 ? "text-rose-400" : "text-white"
                )}>
                  {currentImbalance.toFixed(2)}x
                </span>
              </div>
              <div className="p-3 bg-white/5 rounded-2xl border border-white/5">
                <Database size={20} className="text-slate-500" />
              </div>
            </div>
          </div>

          <div className="md:col-span-2 bg-slate-900/40 backdrop-blur-2xl border border-white/10 rounded-[2.5rem] p-8 shadow-2xl relative overflow-hidden">
            <div className="flex items-center justify-between mb-8">
              <div className="flex items-center gap-2 text-slate-500 font-black text-[10px] tracking-widest uppercase">
                <Gauge size={14} className="text-purple-500" />
                AI Contextual Analysis
              </div>
              <div className="px-3 py-1 bg-purple-500/10 rounded-full border border-purple-500/20 text-[9px] font-black text-purple-400 tracking-widest uppercase flex items-center gap-2">
                <div className="w-1.5 h-1.5 rounded-full bg-purple-500 animate-pulse" />
                Llama-3 GPU
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-8 items-center">
              <div className="text-7xl font-mono font-bold tracking-tighter">
                {currentSentiment > 0 ? '+' : ''}{currentSentiment.toFixed(2)}
              </div>
              <div className="space-y-4">
                <div className="h-4 w-full bg-slate-800/50 rounded-full overflow-hidden p-1 border border-white/5">
                  <div
                    className="h-full bg-gradient-to-r from-rose-500 via-slate-400 to-emerald-500 rounded-full transition-all duration-1000 ease-in-out shadow-[0_0_20px_rgba(59,130,246,0.3)]"
                    style={{ width: `${sentimentPos}%` }}
                  />
                </div>
                <div className="flex justify-between text-[9px] font-black text-slate-500 tracking-widest uppercase">
                  <span>Bearish</span>
                  <span>Neutral</span>
                  <span>Bullish</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
          <div className="md:col-span-3 bg-slate-900/40 backdrop-blur-2xl border border-white/10 rounded-[3rem] p-2 shadow-2xl">
            <div className="p-8">
              <div className="flex items-center justify-between mb-8">
                <div className="flex items-center gap-2 text-slate-500 font-black text-[10px] tracking-widest uppercase">
                  <TrendingUp size={14} className="text-emerald-500" />
                  Live Market Candlesticks (Persistent)
                </div>
                <div className="flex items-center gap-2 px-4 py-2 bg-emerald-500/10 rounded-2xl border border-emerald-500/20">
                  <div className="h-2 w-2 rounded-full bg-emerald-500 animate-ping" />
                  <span className="text-[10px] font-black text-emerald-500 uppercase tracking-widest">HFT STREAM</span>
                </div>
              </div>
              <CandleChart data={candles} />
            </div>
          </div>

          <div className="md:col-span-1 bg-slate-900/40 backdrop-blur-2xl border border-white/10 rounded-[3rem] p-8 shadow-2xl overflow-hidden flex flex-col">
            <div className="flex items-center justify-between mb-6">
              <div className="flex items-center gap-2 text-slate-500 font-black text-[10px] tracking-widest uppercase">
                <RefreshCw size={14} className="text-rose-500 animate-spin-slow" />
                Liquidation Feed
              </div>
              <button 
                onClick={() => {
                  setLiquidations([]);
                  localStorage.removeItem('sentigraph_liquidations');
                }}
                className="p-2 hover:bg-rose-500/10 rounded-xl transition-colors group"
                title="Clear Liquidation Feed"
              >
                <Trash2 size={12} className="text-slate-600 group-hover:text-rose-500" />
              </button>
            </div>
            <div className="space-y-4 overflow-y-auto max-h-[500px] pr-2 custom-scrollbar">
              {liquidations.length === 0 ? (
                <div className="text-center py-12 text-slate-600 italic text-xs">Waiting for liquidations...</div>
              ) : (
                liquidations.map((liq, i) => (
                  <div key={i} className="bg-white/5 border border-white/5 rounded-2xl p-4 flex flex-col gap-1 animate-in fade-in slide-in-from-right-4 duration-500">
                    <div className="flex justify-between items-center">
                      <span className={cn(
                        "text-[10px] font-black px-2 py-0.5 rounded-md uppercase tracking-tighter",
                        liq.side === 'SELL' ? "bg-rose-500/20 text-rose-400" : "bg-emerald-500/20 text-emerald-400"
                      )}>
                        {liq.side === 'SELL' ? 'Short Liquidation' : 'Long Liquidation'}
                      </span>
                      <span className="text-[10px] font-mono text-slate-500">{liq.symbol.toUpperCase()}</span>
                    </div>
                    <div className="flex justify-between items-end mt-1">
                      <div className="text-sm font-bold font-mono tracking-tighter">
                        ${liq.price.toLocaleString()}
                      </div>
                      <div className="text-[10px] font-black text-slate-400">
                        {liq.amount.toFixed(4)} {liq.symbol.toUpperCase()}
                      </div>
                    </div>
                  </div>
                ))
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
