'use client';

import React, { useEffect, useRef } from 'react';
import { createChart, ColorType, ISeriesApi, AreaSeries, Time } from 'lightweight-charts';

interface PriceChartProps {
  data: { time: Time; value: number }[];
}

const PriceChart: React.FC<PriceChartProps> = ({ data }) => {
  const chartContainerRef = useRef<HTMLDivElement>(null);
  const chartRef = useRef<any>(null);
  const seriesRef = useRef<ISeriesApi<"Area"> | null>(null);

  useEffect(() => {
    if (!chartContainerRef.current) return;

    const chart = createChart(chartContainerRef.current, {
      layout: {
        background: { type: ColorType.Solid, color: 'transparent' },
        textColor: '#64748b',
        fontSize: 10,
        fontFamily: 'JetBrains Mono, ui-monospace, monospace',
      },
      grid: {
        vertLines: { color: 'rgba(255, 255, 255, 0.02)' },
        horzLines: { color: 'rgba(255, 255, 255, 0.02)' },
      },
      width: chartContainerRef.current.clientWidth,
      height: 400,
      timeScale: {
        timeVisible: true,
        secondsVisible: true,
        borderColor: 'rgba(255, 255, 255, 0.05)',
      },
      rightPriceScale: {
        borderColor: 'rgba(255, 255, 255, 0.05)',
        scaleMargins: {
          top: 0.1,
          bottom: 0.1,
        },
      },
      crosshair: {
        vertLine: {
          color: '#3b82f6',
          width: 1,
          style: 0,
        },
        horzLine: {
          color: '#3b82f6',
          width: 1,
          style: 0,
        },
      },
      handleScale: {
        mouseWheel: true,
      },
      handleScroll: {
        mouseWheel: true,
      }
    });

    const series = chart.addSeries(AreaSeries, {
      lineColor: '#3b82f6',
      lineWidth: 2,
      topColor: 'rgba(59, 130, 246, 0.25)',
      bottomColor: 'rgba(59, 130, 246, 0.0)',
      priceFormat: {
        type: 'price',
        precision: 2,
        minMove: 0.01,
      },
    });

    chartRef.current = chart;
    seriesRef.current = series;

    const handleResize = () => {
      chart.applyOptions({ width: chartContainerRef.current!.clientWidth });
    };

    window.addEventListener('resize', handleResize);

    return () => {
      window.removeEventListener('resize', handleResize);
      chart.remove();
    };
  }, []);

  useEffect(() => {
    if (seriesRef.current && data.length > 0) {
      seriesRef.current.setData(data);
      // Auto-fit content if it's the first data load
      if (data.length < 201) {
          chartRef.current.timeScale().fitContent();
      }
    }
  }, [data]);

  return (
    <div className="relative">
      <div ref={chartContainerRef} className="w-full" />
    </div>
  );
};

export default PriceChart;
