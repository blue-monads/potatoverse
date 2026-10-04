"use client";

import React, { useState, useEffect, useMemo } from 'react';
import {
    getRootRouting,
    updateRootRouting,
    reloadRootRouting,
    RootRouteTarget,
    Space
} from '@/lib/api';
import { useGApp } from '@/hooks';
import {
    Globe,
    RefreshCw,
    Plus,
    Trash2,
    Save,
    Shield,
    CheckCircle2,
    AlertTriangle,
    Layers,
    Info,
    Search
} from 'lucide-react';

interface RouteEntry {
    id: string; // internal tracking key
    pattern: string;
    spaceId: number;
}

export default function RoutingSettingsPage() {
    const gapp = useGApp();

    const [routeEntries, setRouteEntries] = useState<RouteEntry[]>([]);
    const [originalRoutes, setOriginalRoutes] = useState<Record<string, RootRouteTarget>>({});
    const [mainSpaces, setMainSpaces] = useState<Space[]>([]);

    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [reloading, setReloading] = useState(false);

    const [searchTerm, setSearchTerm] = useState('');
    const [alertMessage, setAlertMessage] = useState<{ type: 'success' | 'error' | 'info'; text: string } | null>(null);

    const loadData = async () => {
        try {
            setLoading(true);
            setAlertMessage(null);
            const res = await getRootRouting();
            const { routes, main_spaces, } = res.data;

            const entries: RouteEntry[] = Object.entries(routes || {}).map(([pattern, target]) => ({
                id: `${pattern}-${target.space_id}-${Math.random()}`,
                pattern,
                spaceId: target.space_id,
            }));

            // Sort entries: specific domains first, then wildcards, then catchall
            entries.sort((a, b) => {
                if (a.pattern === '*') return 1;
                if (b.pattern === '*') return -1;
                return a.pattern.localeCompare(b.pattern);
            });

            setRouteEntries(entries);
            setOriginalRoutes(routes || {});
            setMainSpaces(main_spaces || []);
        } catch (err: any) {
            console.error('Failed to load routing data:', err);
            const msg = err.response?.data?.message || err.message || 'Failed to load routing settings';
            setAlertMessage({ type: 'error', text: msg });
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        if (!gapp.isInitialized) return;
        loadData();
    }, [gapp.isInitialized]);

    // Track if there are unsaved changes
    const isDirty = useMemo(() => {
        const currentMap: Record<string, number> = {};
        for (const r of routeEntries) {
            const p = r.pattern.trim().toLowerCase();
            if (p) {
                currentMap[p] = r.spaceId;
            }
        }

        const origKeys = Object.keys(originalRoutes);
        const currKeys = Object.keys(currentMap);

        if (origKeys.length !== currKeys.length) return true;
        for (const k of origKeys) {
            if (originalRoutes[k]?.space_id !== currentMap[k]) return true;
        }
        return false;
    }, [routeEntries, originalRoutes]);

    const handleAddRoute = () => {
        const defaultSpaceId = mainSpaces.length > 0 ? mainSpaces[0].id : 0;
        setRouteEntries(prev => [
            ...prev,
            {
                id: `new-${Date.now()}-${Math.random()}`,
                pattern: '',
                spaceId: defaultSpaceId,
            }
        ]);
    };

    const handleRemoveRoute = (id: string) => {
        setRouteEntries(prev => prev.filter(r => r.id !== id));
    };

    const handlePatternChange = (id: string, newPattern: string) => {
        setRouteEntries(prev => prev.map(r => r.id === id ? { ...r, pattern: newPattern } : r));
    };

    const handleSpaceChange = (id: string, newSpaceId: number) => {
        setRouteEntries(prev => prev.map(r => r.id === id ? { ...r, spaceId: newSpaceId } : r));
    };

    const handleSave = async () => {
        try {
            setSaving(true);
            setAlertMessage(null);

            // Validation
            const routeMap: Record<string, RootRouteTarget> = {};
            for (const entry of routeEntries) {
                const pattern = entry.pattern.trim().toLowerCase();
                if (!pattern) continue;

                if (!entry.spaceId || entry.spaceId <= 0) {
                    throw new Error(`Please select a valid main space for domain pattern "${pattern}"`);
                }

                // Verify space is in mainSpaces
                const space = mainSpaces.find(s => s.id === entry.spaceId);
                if (!space) {
                    throw new Error(`Space #${entry.spaceId} is not a valid main space. Only main spaces are allowed in root routing.`);
                }

                if (routeMap[pattern]) {
                    throw new Error(`Duplicate domain pattern detected: "${pattern}". Each domain pattern must be unique.`);
                }

                routeMap[pattern] = { space_id: entry.spaceId };
            }

            const res = await updateRootRouting(routeMap);
            setAlertMessage({
                type: 'success',
                text: res.data.message || 'Routing index successfully updated and reloaded on engine!'
            });

            // Refresh state
            setOriginalRoutes(res.data.routes || routeMap);
        } catch (err: any) {
            console.error('Failed to save routing index:', err);
            const msg = err.response?.data?.message || err.message || 'Failed to save routing index';
            setAlertMessage({ type: 'error', text: msg });
        } finally {
            setSaving(false);
        }
    };

    const handleReload = async () => {
        try {
            setReloading(true);
            setAlertMessage(null);
            const res = await reloadRootRouting();
            setAlertMessage({
                type: 'success',
                text: res.data.message || 'Engine root routing index reloaded successfully!'
            });
        } catch (err: any) {
            console.error('Failed to reload routing index:', err);
            const msg = err.response?.data?.message || err.message || 'Failed to reload routing index';
            setAlertMessage({ type: 'error', text: msg });
        } finally {
            setReloading(false);
        }
    };

    const filteredRoutes = useMemo(() => {
        if (!searchTerm.trim()) return routeEntries;
        const q = searchTerm.toLowerCase();
        return routeEntries.filter(r => {
            const space = mainSpaces.find(s => s.id === r.spaceId);
            const spaceName = space?.namespace_key || '';
            return r.pattern.toLowerCase().includes(q) || spaceName.toLowerCase().includes(q);
        });
    }, [routeEntries, searchTerm, mainSpaces]);

    if (loading) {
        return (
            <div className="flex flex-col items-center justify-center p-12 space-y-4">
                <RefreshCw className="w-8 h-8 text-primary-500 animate-spin" />
                <p className="text-gray-500 text-sm">Loading root routing configuration...</p>
            </div>
        );
    }

    return (
        <div className="space-y-6 pb-12">
            {/* Header / Actions Card */}
            <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
                <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
                    <div className="space-y-1">
                        <div className="flex items-center gap-2">
                            <Globe className="w-6 h-6 text-primary-600" />
                            <h2 className="text-2xl font-bold text-gray-900">Root Domain Routing</h2>
                            <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-purple-100 text-purple-800 border border-purple-200">
                                <Shield className="w-3 h-3" />
                                Admin Restricted
                            </span>
                        </div>
                        <p className="text-sm text-gray-600 max-w-2xl">
                            Map incoming domain names and host patterns directly to <strong className="font-semibold text-gray-800">Main Spaces</strong> (primary package spaces).
                            Only main spaces can be selected for root domain routing.
                        </p>
                    </div>

                    <div className="flex items-center gap-2 flex-wrap">
                        <button
                            onClick={handleReload}
                            disabled={reloading || saving}
                            className="inline-flex items-center gap-2 px-3.5 py-2 rounded-lg border border-gray-300 bg-white hover:bg-gray-50 text-gray-700 text-sm font-medium transition shadow-sm disabled:opacity-50"
                            title="Force reload the in-memory routing index in the engine"
                        >
                            <RefreshCw className={`w-4 h-4 ${reloading ? 'animate-spin text-primary-600' : ''}`} />
                            {reloading ? 'Reloading...' : 'Reload Index'}
                        </button>

                        <button
                            onClick={handleSave}
                            disabled={saving || !isDirty}
                            className="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-primary-600 hover:bg-primary-700 text-white text-sm font-medium transition shadow-sm disabled:opacity-50"
                        >
                            <Save className={`w-4 h-4 ${saving ? 'animate-spin' : ''}`} />
                            {saving ? 'Saving...' : isDirty ? 'Save Routing' : 'Saved'}
                        </button>
                    </div>
                </div>

                {/* Alerts */}
                {alertMessage && (
                    <div className={`mt-4 p-4 rounded-lg flex items-center gap-3 text-sm ${alertMessage.type === 'success'
                        ? 'bg-green-50 text-green-800 border border-green-200'
                        : alertMessage.type === 'error'
                            ? 'bg-red-50 text-red-800 border border-red-200'
                            : 'bg-blue-50 text-blue-800 border border-blue-200'
                        }`}>
                        {alertMessage.type === 'success' ? (
                            <CheckCircle2 className="w-5 h-5 flex-shrink-0 text-green-600" />
                        ) : alertMessage.type === 'error' ? (
                            <AlertTriangle className="w-5 h-5 flex-shrink-0 text-red-600" />
                        ) : (
                            <Info className="w-5 h-5 flex-shrink-0 text-blue-600" />
                        )}
                        <span className="flex-1">{alertMessage.text}</span>
                        <button
                            onClick={() => setAlertMessage(null)}
                            className="text-gray-400 hover:text-gray-600"
                        >
                            ✕
                        </button>
                    </div>
                )}

                {/* Status Bar */}
                <div className="mt-4 pt-4 border-t border-gray-100 flex flex-wrap items-center justify-between gap-3 text-xs text-gray-500">
                    <div className="flex items-center gap-4">
                        <span className="flex items-center gap-1.5">
                            <span className={`w-2 h-2 rounded-full ${mainSpaces.length > 0 ? 'bg-green-500' : 'bg-amber-500'}`}></span>
                            Available Main Spaces: <strong className="text-gray-700 font-semibold">{mainSpaces.length}</strong>
                        </span>
                        <span className="flex items-center gap-1.5">
                            <span className="w-2 h-2 rounded-full bg-blue-500"></span>
                            Active Route Rules: <strong className="text-gray-700 font-semibold">{routeEntries.length}</strong>
                        </span>
                    </div>

                    {isDirty && (
                        <span className="px-2 py-0.5 rounded bg-amber-100 text-amber-800 font-medium">
                            Unsaved changes
                        </span>
                    )}
                </div>
            </div>

            {/* Zero Main Spaces Notice */}
            {mainSpaces.length === 0 && (
                <div className="bg-amber-50 border border-amber-200 rounded-xl p-5 flex items-start gap-4">
                    <AlertTriangle className="w-6 h-6 text-amber-600 flex-shrink-0 mt-0.5" />
                    <div>
                        <h4 className="text-sm font-bold text-amber-900">No main spaces found</h4>
                        <p className="text-xs text-amber-700 mt-1 leading-relaxed">
                            Root routing rules map to main spaces (primary spaces for installed apps).
                            Install an app package to configure root domain mappings.
                        </p>
                    </div>
                </div>
            )}

            {/* Routing Rules Table Card */}
            <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
                <div className="p-4 sm:p-6 border-b border-gray-100 flex flex-col sm:flex-row sm:items-center justify-between gap-4 bg-gray-50/50">
                    <div>
                        <h3 className="text-lg font-bold text-gray-900">Domain Mapping Rules</h3>
                        <p className="text-xs text-gray-500">
                            Configure hostnames or wildcards and assign each to a main space.
                        </p>
                    </div>

                    <div className="flex items-center gap-3">
                        <div className="relative">
                            <Search className="w-4 h-4 absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400" />
                            <input
                                type="text"
                                placeholder="Search rules..."
                                value={searchTerm}
                                onChange={(e) => setSearchTerm(e.target.value)}
                                className="pl-9 pr-3 py-1.5 text-xs bg-white border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500 w-48 sm:w-64"
                            />
                        </div>

                        <button
                            onClick={handleAddRoute}
                            disabled={mainSpaces.length === 0}
                            className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-primary-600 hover:bg-primary-700 text-white text-xs font-semibold shadow-sm transition disabled:opacity-50"
                        >
                            <Plus className="w-4 h-4" />
                            Add Rule
                        </button>
                    </div>
                </div>

                {filteredRoutes.length === 0 ? (
                    <div className="text-center py-12 px-4">
                        <Globe className="w-12 h-12 text-gray-300 mx-auto mb-3" />
                        <h4 className="text-sm font-semibold text-gray-700">No routing rules configured</h4>
                        <p className="text-xs text-gray-500 mt-1 max-w-md mx-auto">
                            {searchTerm
                                ? 'No rules match your search filter.'
                                : 'Click "Add Rule" above to create your first root domain routing mapping.'}
                        </p>
                        {mainSpaces.length > 0 && !searchTerm && (
                            <button
                                onClick={handleAddRoute}
                                className="mt-4 inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-primary-50 text-primary-700 text-xs font-semibold hover:bg-primary-100 transition"
                            >
                                <Plus className="w-4 h-4" />
                                Create First Rule
                            </button>
                        )}
                    </div>
                ) : (
                    <div className="overflow-x-auto">
                        <table className="w-full text-left text-sm">
                            <thead className="bg-gray-50 text-xs text-gray-500 font-semibold uppercase tracking-wider border-b border-gray-200">
                                <tr>
                                    <th className="py-3 px-4 w-12 text-center">#</th>
                                    <th className="py-3 px-4">Domain / Host Pattern</th>
                                    <th className="py-3 px-4">Match Type</th>
                                    <th className="py-3 px-4">Target Main Space</th>
                                    <th className="py-3 px-4 w-20 text-center">Action</th>
                                </tr>
                            </thead>
                            <tbody className="divide-y divide-gray-100">
                                {filteredRoutes.map((route, idx) => {
                                    const p = route.pattern.trim().toLowerCase();
                                    const isCatchAll = p === '*';
                                    const isWildcard = p.startsWith('*.');
                                    const matchTypeBadge = isCatchAll ? (
                                        <span className="px-2 py-0.5 rounded text-[11px] font-semibold bg-red-100 text-red-800">
                                            Catch-All Fallback
                                        </span>
                                    ) : isWildcard ? (
                                        <span className="px-2 py-0.5 rounded text-[11px] font-semibold bg-amber-100 text-amber-800">
                                            Wildcard Subdomain
                                        </span>
                                    ) : p ? (
                                        <span className="px-2 py-0.5 rounded text-[11px] font-semibold bg-blue-100 text-blue-800">
                                            Exact Hostname
                                        </span>
                                    ) : (
                                        <span className="px-2 py-0.5 rounded text-[11px] font-semibold bg-gray-100 text-gray-500">
                                            Pending input
                                        </span>
                                    );

                                    return (
                                        <tr key={route.id} className="hover:bg-gray-50/70 transition">
                                            <td className="py-3 px-4 text-center text-xs text-gray-400 font-mono">
                                                {idx + 1}
                                            </td>

                                            <td className="py-3 px-4">
                                                <div className="space-y-1">
                                                    <input
                                                        type="text"
                                                        value={route.pattern}
                                                        onChange={(e) => handlePatternChange(route.id, e.target.value)}
                                                        placeholder="e.g. example.com or *.company.com or *"
                                                        className="w-full max-w-sm px-3 py-1.5 text-xs font-mono bg-white border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500 text-gray-900"
                                                    />
                                                </div>
                                            </td>

                                            <td className="py-3 px-4">
                                                {matchTypeBadge}
                                            </td>

                                            <td className="py-3 px-4">
                                                <div className="flex items-center gap-2 max-w-xs">
                                                    <Layers className="w-4 h-4 text-purple-600 flex-shrink-0" />
                                                    <select
                                                        value={route.spaceId || 0}
                                                        onChange={(e) => handleSpaceChange(route.id, parseInt(e.target.value, 10))}
                                                        className="w-full px-2.5 py-1.5 text-xs bg-white border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500 text-gray-900 font-medium"
                                                    >
                                                        {mainSpaces.length === 0 ? (
                                                            <option value="0" disabled>No main spaces available</option>
                                                        ) : (
                                                            mainSpaces.map(s => (
                                                                <option key={s.id} value={s.id}>
                                                                    {s.namespace_key} (Space #{s.id})
                                                                </option>
                                                            ))
                                                        )}
                                                    </select>
                                                </div>
                                            </td>

                                            <td className="py-3 px-4 text-center">
                                                <button
                                                    onClick={() => handleRemoveRoute(route.id)}
                                                    className="p-1.5 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition"
                                                    title="Remove rule"
                                                >
                                                    <Trash2 className="w-4 h-4" />
                                                </button>
                                            </td>
                                        </tr>
                                    );
                                })}
                            </tbody>
                        </table>
                    </div>
                )}
            </div>

            {/* Pattern Documentation & Matching Logic Reference */}
            <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
                <div className="flex items-center gap-2 mb-3">
                    <Info className="w-5 h-5 text-primary-600" />
                    <h4 className="text-base font-bold text-gray-900">How Root Routing Works</h4>
                </div>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-xs text-gray-600">
                    <div className="p-3.5 bg-gray-50 rounded-lg border border-gray-100">
                        <strong className="text-gray-900 font-semibold block mb-1">1. Exact Hostname Match</strong>
                        <p>
                            Matches full domain or host from the request.
                            <br />
                            <code className="bg-white px-1 py-0.5 rounded border border-gray-200 font-mono text-purple-700 mt-1 inline-block">
                                myapp.example.com
                            </code>
                        </p>
                    </div>

                    <div className="p-3.5 bg-gray-50 rounded-lg border border-gray-100">
                        <strong className="text-gray-900 font-semibold block mb-1">2. Wildcard Prefix Fallback</strong>
                        <p>
                            Prefix labels are progressively stripped:
                            <br />
                            <code className="bg-white px-1 py-0.5 rounded border border-gray-200 font-mono text-purple-700 mt-1 inline-block">
                                *.example.com
                            </code>
                            <br />
                            Matches <code className="font-mono">foo.example.com</code> and <code className="font-mono">example.com</code>.
                        </p>
                    </div>

                    <div className="p-3.5 bg-gray-50 rounded-lg border border-gray-100">
                        <strong className="text-gray-900 font-semibold block mb-1">3. Global Catch-All Fallback</strong>
                        <p>
                            Use <code className="bg-white px-1 py-0.5 rounded border border-gray-200 font-mono text-purple-700 font-bold">*</code> to route any domain that doesn&apos;t match an exact or wildcard rule. Requests with no match return <strong>404</strong>.
                        </p>
                    </div>
                </div>
            </div>
        </div>
    );
}
