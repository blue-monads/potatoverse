"use client";
import React from 'react';

interface TabViewProps {
    activeView: 'tables' | 'livequery';
    onViewChange: (view: 'tables' | 'livequery') => void;
    TablesViewComponent: React.ComponentType<any>;
    tablesProps: any;
    LiveQueryViewComponent: React.ComponentType<any>;
    liveQueryProps?: any;
}

export default function TabView({
    activeView,
    onViewChange,
    TablesViewComponent,
    tablesProps,
    LiveQueryViewComponent,
    liveQueryProps,
}: TabViewProps) {
    return (
        <div className="max-w-7xl mx-auto px-6 py-8 w-full flex flex-col flex-1 overflow-hidden">
            {/* View Switcher */}
            <div className="flex justify-between items-center mb-4">
                <h2 className="text-xl font-semibold text-gray-900">
                    {activeView === 'tables' ? 'Tables' : 'Live Query'}
                </h2>
                <div className="inline-flex bg-gray-100 rounded-lg p-1">
                    <button
                        onClick={() => onViewChange('tables')}
                        className={`px-4 py-2 rounded-md text-sm font-medium transition-all ${activeView === 'tables'
                                ? 'bg-white text-gray-900 shadow-sm'
                                : 'text-gray-600 hover:text-gray-900 hover:bg-gray-200'
                            }`}
                    >
                        Tables
                    </button>
                    <button
                        onClick={() => onViewChange('livequery')}
                        className={`px-4 py-2 rounded-md text-sm font-medium transition-all ${activeView === 'livequery'
                                ? 'bg-white text-gray-900 shadow-sm'
                                : 'text-gray-600 hover:text-gray-900 hover:bg-gray-200'
                            }`}
                    >
                        Live Query
                    </button>
                </div>
            </div>

            {activeView === 'tables' ? (
                <TablesViewComponent {...tablesProps} />
            ) : (
                <LiveQueryViewComponent {...liveQueryProps} />
            )}
        </div>
    );
}
