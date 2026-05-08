import { useEffect, useState, useRef } from 'react';

export interface MarketData {
  price: string;
  sentiment: number;
  updated_at: string;
}

export const useMarketData = (url: string) => {
  const [data, setData] = useState<MarketData | null>(null);
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
          const parsedData = JSON.parse(event.data);
          const processedData = {
            ...parsedData,
            sentiment: typeof parsedData.sentiment === 'string'
              ? parseFloat(parsedData.sentiment)
              : parsedData.sentiment,
            updated_at: parsedData.timestamp || parsedData.updated_at,
          };
          setData(processedData);
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
