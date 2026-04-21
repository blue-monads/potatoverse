"use client";
import React, { useState } from 'react';
import { Database, Play, Copy, RefreshCw } from 'lucide-react';
import { useParams, useSearchParams } from 'next/navigation';
import { sqlQuerySpaceData } from '@/lib';

export default function LiveQueryViewContent() {
    const sparams = useSearchParams();
    const installId = sparams.get('install_id') as string;
    const [query, setQuery] = useState('SELECT * FROM users LIMIT 10;');
    const [results, setResults] = useState<any[]>([]);
    const [isLoading, setIsLoading] = useState(false);
    const [columns, setColumns] = useState<string[]>([]);
    const [error, setError] = useState<string | null>(null);

    const handleRunQuery = async () => {
        if (!installId) {
            setError('Install ID not available');
            console.log("@params", sparams);
            return;
        }

        setIsLoading(true);
        setError(null);

        try {
             const resp = await sqlQuerySpaceData(Number(installId), query);
             if (resp.status !== 200) {
                setError(`Error: ${resp.statusText}, ${resp.data?.error || 'Unknown error'}`);

                return;
             }

             const data = await resp.data;
                if (data.length > 0) {
                    setColumns(Object.keys(data[0]));
                    setResults(data);
                } else {
                    setColumns([]);
                    setResults([]);
                }
        } catch (err: any) {
            setError(err.message || 'An error occurred while running the query');
            setResults([]);
            setColumns([]);

        }


        
    };

    return (
        <div className="bg-white rounded-lg shadow border border-gray-200 overflow-hidden flex flex-col h-full">
            {/* Editor Section */}
            <div className="flex-1 flex flex-col border-b border-gray-200">
                <div className="bg-gray-50 px-4 py-3 border-b border-gray-200 flex items-center justify-between">
                    <div className="flex items-center gap-2">
                        <Database className="w-4 h-4 text-gray-600" />
                        <span className="text-sm font-medium text-gray-700">SQL Query Editor</span>
                    </div>
                    <button
                        onClick={() => setQuery('')}
                        className="text-xs text-gray-500 hover:text-gray-700"
                    >
                        Clear
                    </button>
                </div>
                
                <textarea
                    value={query}
                    onChange={(e) => setQuery(e.target.value)}
                    placeholder="Enter your SQL query here..."
                    className="p-4 font-mono text-sm resize-y focus:outline-none "
                    spellCheck="false"
                />

                <div className="bg-gray-50 px-4 py-3 flex gap-2 border-t border-gray-200">
                    <button
                        onClick={handleRunQuery}
                        disabled={isLoading || !query.trim()}
                        className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed text-sm font-medium"
                    >
                        <Play className="w-4 h-4" />
                        Run Query
                    </button>
                    <button
                        onClick={() => navigator.clipboard.writeText(query)}
                        className="flex items-center gap-2 px-3 py-2 bg-gray-200 text-gray-700 rounded hover:bg-gray-300 text-sm"
                    >
                        <Copy className="w-4 h-4" />
                    </button>
                    <button
                        onClick={() => handleRunQuery()}
                        className="flex items-center gap-2 px-3 py-2 bg-gray-200 text-gray-700 rounded hover:bg-gray-300 text-sm"
                    >
                        <RefreshCw className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`} />
                    </button>
                </div>
            </div>

            {/* Results Section */}
            <div className="flex-1 flex flex-col overflow-hidden">
                <div className="bg-gray-50 px-4 py-3 border-b border-gray-200">
                    <span className="text-sm font-medium text-gray-700">
                        Results {results.length > 0 && `(${results.length} rows)`}
                    </span>
                </div>

                {error ? (
                    <div className="flex-1 flex items-center justify-center">
                        <div className="text-center">
                            <div className="text-red-600 text-sm font-medium mb-2">Error</div>
                            <p className="text-red-500 text-xs max-w-md">{error}</p>
                        </div>
                    </div>
                ) : results.length === 0 ? (
                    <div className="flex-1 flex items-center justify-center text-gray-500">
                        <div className="text-center">
                            <Database className="w-12 h-12 mx-auto mb-3 opacity-30" />
                            <p className="text-sm">Run a query to see results</p>
                        </div>
                    </div>
                ) : (
                    <div className="flex-1 overflow-auto">
                        <table className="w-full text-sm">
                            <thead className="bg-gray-100 sticky top-0">
                                <tr>
                                    {columns.map((col) => (
                                        <th
                                            key={col}
                                            className="px-4 py-2 text-left font-medium text-gray-700 border-b border-gray-200"
                                        >
                                            {col}
                                        </th>
                                    ))}
                                </tr>
                            </thead>
                            <tbody>
                                {results.map((row, idx) => (
                                    <tr key={idx} className="border-b border-gray-200 hover:bg-gray-50">
                                        {columns.map((col) => (
                                            <td
                                                key={`${idx}-${col}`}
                                                className="px-4 py-2 text-gray-800"
                                            >
                                                {row[col]}
                                            </td>
                                        ))}
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                )}
            </div>
        </div>
    );
}
