"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import { Smartphone, ArrowLeft, Monitor, Plus, X, Copy, Check } from "lucide-react";
import { getSelfDevices, createSelfDevice, UserDevice, CreateDeviceResponse } from "@/lib/api";
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
    const [addOpen, setAddOpen] = useState(false);
    const [addName, setAddName] = useState("");
    const [addSubmitting, setAddSubmitting] = useState(false);
    const [addError, setAddError] = useState<string | null>(null);
    const [created, setCreated] = useState<CreateDeviceResponse | null>(null);
    const [copied, setCopied] = useState(false);

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

    const openAdd = () => {
        setAddOpen(true);
        setAddName("");
        setAddError(null);
        setCreated(null);
        setCopied(false);
    };

    const closeAdd = () => {
        setAddOpen(false);
        if (created) loadDevices();
    };

    const submitAdd = async () => {
        const name = addName.trim();
        if (!name) {
            setAddError("Name is required.");
            return;
        }
        try {
            setAddSubmitting(true);
            setAddError(null);
            const res = await createSelfDevice(name);
            setCreated(res.data);
        } catch (e: unknown) {
            console.error("Failed to create device:", e);
            setAddError(e && typeof e === "object" && "response" in e && e.response && typeof (e.response as { data?: { message?: string } }).data?.message === "string"
                ? (e.response as { data: { message: string } }).data.message
                : "Failed to create device.");
        } finally {
            setAddSubmitting(false);
        }
    };

    const copyToken = async () => {
        if (!created?.token) return;
        try {
            await navigator.clipboard.writeText(created.token);
            setCopied(true);
            setTimeout(() => setCopied(false), 2000);
        } catch {
            setAddError("Could not copy to clipboard.");
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

                        <div className="flex items-center gap-2">
                            <button
                                type="button"
                                onClick={openAdd}
                                className="btn btn-base preset-filled bg-primary-600 text-white"
                            >
                                <Plus className="w-4 h-4" />
                                Add Device
                            </button>
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

            {addOpen && (
                <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50">
                    <div className="bg-white rounded-xl border border-gray-200 shadow-xl max-w-md w-full max-h-[90vh] overflow-y-auto">
                        <div className="flex items-center justify-between p-4 border-b border-gray-200">
                            <h2 className="text-lg font-semibold">
                                {created ? "Device created" : "Add device"}
                            </h2>
                            <button
                                type="button"
                                onClick={closeAdd}
                                className="p-1 rounded hover:bg-gray-100"
                                aria-label="Close"
                            >
                                <X className="w-5 h-5" />
                            </button>
                        </div>
                        <div className="p-4 space-y-4">
                            {created ? (
                                <>
                                    <p className="text-sm text-amber-700 bg-amber-50 border border-amber-200 rounded-lg p-3">
                                        Copy the device token now. It won’t be shown again. Use it with <code className="text-xs bg-amber-100 px-1 rounded">POST /zz/api/core/auth/device-token</code> to get an access token.
                                    </p>
                                    <div className="flex items-center gap-2">
                                        <input
                                            type="text"
                                            readOnly
                                            value={created.token}
                                            className="flex-1 font-mono text-sm p-2 border border-gray-200 rounded bg-gray-50"
                                        />
                                        <button
                                            type="button"
                                            onClick={copyToken}
                                            className="btn btn-base preset-filled bg-primary-600 text-white flex items-center gap-1"
                                        >
                                            {copied ? <Check className="w-4 h-4" /> : <Copy className="w-4 h-4" />}
                                            {copied ? "Copied" : "Copy"}
                                        </button>
                                    </div>
                                    {created.expires_on && (
                                        <p className="text-xs text-gray-500">
                                            Expires: {formatDate(created.expires_on)}
                                        </p>
                                    )}
                                </>
                            ) : (
                                <>
                                    {addError && (
                                        <p className="text-sm text-red-600 bg-red-50 border border-red-200 rounded p-2">{addError}</p>
                                    )}
                                    <div>
                                        <label className="block text-sm font-medium text-gray-700 mb-1">Name</label>
                                        <input
                                            type="text"
                                            value={addName}
                                            onChange={(e) => setAddName(e.target.value)}
                                            placeholder="e.g. My API client"
                                            className="w-full p-2 border border-gray-300 rounded-lg"
                                        />
                                    </div>
                                </>
                            )}
                        </div>
                        {!created ? (
                            <div className="flex justify-end gap-2 p-4 border-t border-gray-200">
                                <button
                                    type="button"
                                    onClick={closeAdd}
                                    className="btn btn-base preset-tonal"
                                >
                                    Cancel
                                </button>
                                <button
                                    type="button"
                                    onClick={submitAdd}
                                    disabled={addSubmitting}
                                    className="btn btn-base preset-filled bg-primary-600 text-white disabled:opacity-50"
                                >
                                    {addSubmitting ? "Creating…" : "Create"}
                                </button>
                            </div>
                        ) : (
                            <div className="p-4 border-t border-gray-200">
                                <button
                                    type="button"
                                    onClick={closeAdd}
                                    className="btn btn-base preset-filled bg-primary-600 text-white w-full"
                                >
                                    Done
                                </button>
                            </div>
                        )}
                    </div>
                </div>
            )}
        </div>
    );
}
