"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import { Smartphone, ArrowLeft, Monitor } from "lucide-react";
import { getSelfDevices, UserDevice } from "@/lib/api";
import { useGApp } from "@/hooks";

function formatDate(iso: string) {
    if (!iso) return "—";
    try {
        const d = new Date(iso);
        return d.toLocaleString();
    } catch {
        return iso;
    }
}

export default function Page() {
    return <DevicesPage />;
}

function DevicesPage() {
    const { loaded, isInitialized, isAuthenticated } = useGApp();
    const [devices, setDevices] = useState<UserDevice[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        if (loaded && isInitialized && isAuthenticated) {
            loadDevices();
        }
    }, [loaded, isInitialized, isAuthenticated]);

    const loadDevices = async () => {
        try {
            setLoading(true);
            setError(null);
            const res = await getSelfDevices();
            setDevices(Array.isArray(res.data) ? res.data : []);
        } catch (e) {
            console.error("Failed to load devices:", e);
            setError("Failed to load devices.");
        } finally {
            setLoading(false);
        }
    };

    if (!loaded || !isInitialized) {
        return (
            <div className="min-h-screen bg-gray-50 flex items-center justify-center">
                <div className="text-center">
                    <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4" />
                    <p className="text-gray-600">Initializing...</p>
                </div>
            </div>
        );
    }

    if (!isAuthenticated) {
        return (
            <div className="min-h-screen bg-gray-50 flex items-center justify-center">
                <div className="text-center">
                    <p className="text-gray-600">Please log in to view your devices.</p>
                </div>
            </div>
        );
    }

    return (
        <div className="min-h-screen bg-gray-50">
            <header className="bg-white border-b border-gray-200 px-6 py-4">
                <div className="max-w-4xl mx-auto flex items-center justify-between">
                    <div className="flex items-center gap-4 justify-between w-full">

                        <div className="flex items-center gap-2">
                            <Smartphone className="w-6 h-6 text-blue-600" />
                            <div>
                                <h1 className="text-xl font-bold">Devices</h1>
                                <p className="text-sm text-gray-600">Sessions and devices where you’re signed in</p>
                            </div>
                        </div>


                    </div>
                </div>
            </header>

            <div className="max-w-4xl mx-auto px-6 py-8">
                <div className="bg-white rounded-xl border border-gray-200 overflow-hidden">
                    {loading ? (
                        <div className="p-12 text-center">
                            <div className="animate-spin rounded-full h-10 w-10 border-b-2 border-blue-600 mx-auto mb-4" />
                            <p className="text-gray-600">Loading devices...</p>
                        </div>
                    ) : error ? (
                        <div className="p-12 text-center">
                            <p className="text-red-600 mb-4">{error}</p>
                            <button
                                onClick={loadDevices}
                                className="text-blue-600 hover:underline"
                            >
                                Retry
                            </button>
                        </div>
                    ) : devices.length === 0 ? (
                        <div className="p-12 text-center text-gray-500">
                            <Monitor className="w-12 h-12 mx-auto mb-3 opacity-50" />
                            <p>No devices or sessions yet.</p>
                            <p className="text-sm mt-1">New sessions appear here after you log in.</p>
                        </div>
                    ) : (
                        <div className="overflow-x-auto">
                            <table className="w-full">
                                <thead className="bg-gray-50 border-b border-gray-200">
                                    <tr>
                                        <th className="text-left py-3 px-4 text-sm font-semibold text-gray-700">Name</th>
                                        <th className="text-left py-3 px-4 text-sm font-semibold text-gray-700">Type</th>
                                        <th className="text-left py-3 px-4 text-sm font-semibold text-gray-700">Last IP</th>
                                        <th className="text-left py-3 px-4 text-sm font-semibold text-gray-700">Last login</th>
                                        <th className="text-left py-3 px-4 text-sm font-semibold text-gray-700">Created</th>
                                    </tr>
                                </thead>
                                <tbody className="divide-y divide-gray-100">
                                    {devices.map((d) => (
                                        <tr key={d.id} className="hover:bg-gray-50/50">
                                            <td className="py-3 px-4 text-gray-900 font-medium">{d.name || "—"}</td>
                                            <td className="py-3 px-4 text-gray-600 text-sm">{d.dtype || "session"}</td>
                                            <td className="py-3 px-4 text-gray-600 text-sm font-mono">{d.last_ip || "—"}</td>
                                            <td className="py-3 px-4 text-gray-600 text-sm">{formatDate(d.last_login)}</td>
                                            <td className="py-3 px-4 text-gray-600 text-sm">{formatDate(d.created_at)}</td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
}
