'use client';

import { useState, useEffect } from 'react';

interface Gradient {
  Start: string;
  End: string;
}

export default function GradientsPage() {
  const [gradients, setGradients] = useState<Gradient[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchGradients();
  }, []);

  const fetchGradients = async () => {
    try {
      setLoading(true);
      const response = await fetch('/zz/api/gradients');
      if (!response.ok) {
        throw new Error('Failed to fetch gradients');
      }
      const data = await response.json();
      setGradients(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setLoading(false);
    }
  };

  const getGradientStyle = (gradient: Gradient) => ({
    background: `linear-gradient(135deg, ${gradient.Start}, ${gradient.End})`,
  });

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading gradients...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="text-red-500 text-6xl mb-4">⚠️</div>
          <h2 className="text-2xl font-bold text-gray-800 mb-2">Error Loading Gradients</h2>
          <p className="text-gray-600 mb-4">{error}</p>
          <button 
            onClick={fetchGradients}
            className="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 transition-colors"
          >
            Try Again
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <div className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <h1 className="text-3xl font-bold text-gray-900">Gradient Collection</h1>
          <p className="text-gray-600 mt-2">
            {gradients.length} beautiful gradients from your backend
          </p>
        </div>
      </div>

      {/* Gradient Grid */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4">
          {gradients.map((gradient, index) => (
            <div key={index} className="relative">
              <div 
                className="h-24 rounded-lg shadow-md"
                style={getGradientStyle(gradient)}
              >
              </div>
              <div className="mt-2 text-center">
                <div className="text-xs text-gray-500">#{index + 1}</div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
