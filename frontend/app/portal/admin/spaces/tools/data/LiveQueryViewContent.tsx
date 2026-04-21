"use client";
import React from 'react';
import { Database } from 'lucide-react';

export default function LiveQueryViewContent() {
    return (
        <div className="bg-white rounded-lg shadow overflow-hidden">
            <div className="flex items-center justify-center h-96 text-gray-500">
                <div className="text-center">
                    <Database className="w-12 h-12 mx-auto mb-3 opacity-30" />
                    <p className="text-sm">Live Query functionality coming soon</p>
                </div>
            </div>
        </div>
    );
}
