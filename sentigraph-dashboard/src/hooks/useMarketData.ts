import { useEffect, useState, useRef } from 'react';

export interface MarketData {
  type?: 'market';
  symbol: string;
  price: string;
  sentiment: number;
  imbalance: number;
  updated_at: string;
}

export interface LiquidationData {
  type: 'liquidation';
  symbol: string;
  side: 'BUY' | 'SELL';
  price: number;
  amount: number;
}

export type WSMessage = MarketData | LiquidationData;

export const useMarketData = (url: string) => {
  const [data, setData] = useState<WSMessage | null>(null);
  const [status, setStatus] = useState<'connecting' | 'open' | 'closed'>('connecting');
  const ws = useRef<WebSocket | null>(null);

  useEffect(() => {
    const connect = () => {
      ws.current = new WebSocket(url);

      ws.current.onopen = () => {
        console.log('Connected to SentiGraph-Analyst');
        setStatus('open');
      };

      ws.current.onmessage = (event) => {
        try {
          const parsed = JSON.parse(event.data);

          if (parsed.type === 'liquidation') {
            setData(parsed as LiquidationData);
          } else {
            const processed: MarketData = {
              type: 'market',
              symbol: parsed.symbol || 'btc',
              price: parsed.price,
              sentiment: typeof parsed.sentiment === 'string'
                ? parseFloat(parsed.sentiment)
                : (parsed.sentiment || 0),
              imbalance: parsed.imbalance || 1.0,
              updated_at: parsed.timestamp || parsed.updated_at || new Date().toISOString(),
            };
            setData(processed);
          }
        } catch (error) {
          console.error('Failed to parse market data', error);
        }
      };

      ws.current.onclose = () => {
        console.log('Disconnected from SentiGraph-Analyst');
        setStatus('closed');
        setTimeout(connect, 3000);
      };
    };

    connect();

    return () => {
      if (ws.current) {
        ws.current.close();
      }
    };
  }, [url]);

  return { data, status };
};
