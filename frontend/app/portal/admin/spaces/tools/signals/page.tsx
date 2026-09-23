"use client";
import React, { useState } from 'react';
import { Zap, Edit, Trash2, Send, Eye, RefreshCw, ArrowUpRight, ArrowDownLeft } from 'lucide-react';
import { useRouter, useSearchParams } from 'next/navigation';
import WithAdminBodyLayout from '@/contain/Layouts/WithAdminBodyLayout';
import BigSearchBar from '@/contain/compo/BigSearchBar';
import { AddButton } from '@/contain/AddButton';
import {
    listSignals,
    Signal,
    deleteSignal,
    listSignalEvents,
    SignalTarget,
    updateSignalTargetStatus,
    emitSignal,
} from '@/lib';
import useSimpleDataLoader from '@/hooks/useSimpleDataLoader';

export default function Page() {
    const searchParams = useSearchParams();
    const installId = searchParams.get('install_id');
    const spaceId = searchParams.get('space_id');

    if (!installId) {
        return <div className="p-8 text-center text-gray-500">Install ID not provided</div>;
    }

    return (
        <SignalsListingPage
            installId={parseInt(installId, 10)}
            spaceId={spaceId ? parseInt(spaceId, 10) : undefined}
        />
    );
}

const SignalsListingPage = ({ installId, spaceId }: { installId: number; spaceId?: number }) => {
    const router = useRouter();
    const searchParams = useSearchParams();
    const [searchTerm, setSearchTerm] = useState('');
    const [activeView, setActiveView] = useState<'signals' | 'queue'>('signals');
    const [roleFilter, setRoleFilter] = useState<'all' | 'emitter' | 'receiver'>('all');
    const [statusFilter, setStatusFilter] = useState<string>('all');
    const [selectedPayload, setSelectedPayload] = useState<string | null>(null);

    const signalsLoader = useSimpleDataLoader<Signal[]>({
        loader: () => {
            const role = roleFilter === 'all' ? undefined : roleFilter;
            return listSignals(installId, spaceId, role);
        },
        ready: activeView === 'signals',
        dependencies: [installId, spaceId, activeView, roleFilter],
    });

    const queueLoader = useSimpleDataLoader<SignalTarget[]>({
        loader: () => {
            const status = statusFilter === 'all' ? undefined : statusFilter;
            return listSignalEvents(installId, undefined, status);
        },
        ready: activeView === 'queue',
        dependencies: [installId, spaceId, activeView, statusFilter],
    });

    const filteredSignals = signalsLoader.data?.filter(sig => {
        if (!searchTerm) return true;
        const term = searchTerm.toLowerCase();
        return (
            sig.signal_key.toLowerCase().includes(term) ||
            sig.receiver_handler.toLowerCase().includes(term) ||
            sig.managed_by.toLowerCase().includes(term) ||
            sig.receiver_space_id.toString().includes(term)
        );
    }) || [];

    const filteredQueue = queueLoader.data?.filter(evt => {
        if (!searchTerm) return true;
        const term = searchTerm.toLowerCase();
        return (
            evt.id.toString().includes(term) ||
            evt.signal_id.toString().includes(term) ||
            evt.signal_event_id.toString().includes(term) ||
            (evt.signal_key && evt.signal_key.toLowerCase().includes(term)) ||
            (evt.receiver_handler && evt.receiver_handler.toLowerCase().includes(term)) ||
            evt.status.toLowerCase().includes(term) ||
            (evt.error && evt.error.toLowerCase().includes(term))
        );
    }) || [];

    const handleUnblock = async (targetId: number) => {
        try {
            await updateSignalTargetStatus(installId, targetId, 'new');
            queueLoader.reload();
        } catch (error) {
            console.error('Failed to unblock target:', error);
            alert('Failed to unblock target: ' + ((error as any)?.response?.data?.error || (error as any)?.message));
        }
    };

    const handleDelete = async (id: number) => {
        try {
            await deleteSignal(installId, id);
            signalsLoader.reload();
        } catch (error) {
            console.error('Failed to delete signal:', error);
            alert('Failed to delete signal: ' + ((error as any)?.response?.data?.error || (error as any)?.message));
        }
    };

    const handleEdit = (id: number) => {
        const params = new URLSearchParams(searchParams.toString());
        params.set('signal_id', id.toString());
        router.push(`/portal/admin/spaces/tools/signals/edit?${params.toString()}`);
    };

    const handleNew = () => {
        const params = new URLSearchParams(searchParams.toString());
        router.push(`/portal/admin/spaces/tools/signals/new?${params.toString()}`);
    };

    const handleTestEmit = async (sig: Signal) => {
        const sample = prompt(`Emit test signal for "${sig.signal_key}" with JSON payload:`, JSON.stringify({ test: true, timestamp: Date.now() }));
        if (!sample) return;

        try {
            let parsedPayload: any = sample;
            try {
                parsedPayload = JSON.parse(sample);
            } catch (e) {
                // Keep as raw string
            }

            await emitSignal(installId, {
                signal_key: sig.signal_key,
                emitter_space_id: sig.emitter_space_id,
                payload: parsedPayload,
            });

            alert('Signal emitted successfully! Switch to Queue tab to see processing state.');
            if (activeView === 'queue') {
                queueLoader.reload();
            }
        } catch (error) {
            console.error('Failed to emit test signal:', error);
            alert('Failed to emit test signal: ' + ((error as any)?.response?.data?.error || (error as any)?.message));
        }
    };

    return (
        <WithAdminBodyLayout
            Icon={Zap}
            name="Signals"
            description="Simplified app-to-app signal routing and event delivery"
            variant="none"
        >
            <div className="max-w-7xl mx-auto px-4 py-4 w-full flex flex-col gap-4">
                {/* Header view switcher */}
                <div className="flex justify-between items-center mb-2">
                    <div>
                        <h2 className="text-xl font-semibold text-gray-900">
                            {activeView === 'signals' ? 'Connected Signals' : 'Signal Event Queue'}
                        </h2>
                        <p className="text-xs text-gray-500 mt-0.5">
                            {activeView === 'signals' 
                                ? 'Direct connections between this app and other spaces' 
                                : 'Execution and delivery history of emitted signals'}
                        </p>
                    </div>

                    <div className="inline-flex bg-gray-100 rounded-lg p-1">
                        <button
                            onClick={() => setActiveView('signals')}
                            className={`px-4 py-2 rounded-md text-sm font-medium transition-all ${
                                activeView === 'signals'
                                    ? 'bg-white text-gray-900 shadow-sm'
                                    : 'text-gray-600 hover:text-gray-900 hover:bg-gray-200'
                            }`}
                        >
                            Signals
                        </button>
                        <button
                            onClick={() => setActiveView('queue')}
                            className={`px-4 py-2 rounded-md text-sm font-medium transition-all ${
                                activeView === 'queue'
                                    ? 'bg-white text-gray-900 shadow-sm'
                                    : 'text-gray-600 hover:text-gray-900 hover:bg-gray-200'
                            }`}
                        >
                            Queue
                        </button>
                    </div>
                </div>

                {/* Sub-filters */}
                {activeView === 'signals' ? (
                    <div className="flex gap-2 items-center">
                        <span className="text-xs font-medium text-gray-500 mr-1">Role:</span>
                        <button
                            onClick={() => setRoleFilter('all')}
                            className={`px-3 py-1 text-xs rounded-full border transition-colors ${
                                roleFilter === 'all'
                                    ? 'bg-blue-600 text-white border-blue-600'
                                    : 'bg-white text-gray-600 border-gray-300 hover:bg-gray-50'
                            }`}
                        >
                            All
                        </button>
                        <button
                            onClick={() => setRoleFilter('emitter')}
                            className={`px-3 py-1 text-xs rounded-full border transition-colors flex items-center gap-1 ${
                                roleFilter === 'emitter'
                                    ? 'bg-blue-600 text-white border-blue-600'
                                    : 'bg-white text-gray-600 border-gray-300 hover:bg-gray-50'
                            }`}
                        >
                            <ArrowUpRight className="w-3 h-3" /> Outgoing (Emitter)
                        </button>
                        <button
                            onClick={() => setRoleFilter('receiver')}
                            className={`px-3 py-1 text-xs rounded-full border transition-colors flex items-center gap-1 ${
                                roleFilter === 'receiver'
                                    ? 'bg-blue-600 text-white border-blue-600'
                                    : 'bg-white text-gray-600 border-gray-300 hover:bg-gray-50'
                            }`}
                        >
                            <ArrowDownLeft className="w-3 h-3" /> Incoming (Receiver)
                        </button>
                    </div>
                ) : (
                    <div className="flex gap-2 items-center">
                        <span className="text-xs font-medium text-gray-500 mr-1">Status:</span>
                        {['all', 'new', 'scheduled', 'delayed', 'blocked', 'processed', 'failed', 'expired'].map(st => (
                            <button
                                key={st}
                                onClick={() => setStatusFilter(st)}
                                className={`px-3 py-1 text-xs rounded-full border capitalize transition-colors ${
                                    statusFilter === st
                                        ? 'bg-blue-600 text-white border-blue-600'
                                        : 'bg-white text-gray-600 border-gray-300 hover:bg-gray-50'
                                }`}
                            >
                                {st}
                            </button>
                        ))}
                    </div>
                )}

                {/* Search Bar & Add Button */}
                <BigSearchBar
                    searchText={searchTerm}
                    setSearchText={setSearchTerm}
                    rightContent={
                        activeView === 'signals' && (
                            <AddButton
                                name="+ New Signal"
                                onClick={handleNew}
                            />
                        )
                    }
                />

                {/* Signals Table */}
                {activeView === 'signals' ? (
                    <div className="bg-white rounded-lg shadow overflow-hidden border border-gray-200">
                        <div className="overflow-x-auto">
                            <table className="min-w-full divide-y divide-gray-200">
                                <thead className="bg-gray-50">
                                    <tr>
                                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                            Signal Key
                                        </th>
                                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                            Emitter Space
                                        </th>
                                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                            Receiver Space & Handler
                                        </th>
                                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                            Retries / Delay
                                        </th>
                                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                            Status
                                        </th>
                                        <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                                            Actions
                                        </th>
                                    </tr>
                                </thead>
                                <tbody className="bg-white divide-y divide-gray-200">
                                    {signalsLoader.loading ? (
                                        <tr>
                                            <td colSpan={6} className="px-6 py-6 text-center text-gray-500">
                                                Loading signals...
                                            </td>
                                        </tr>
                                    ) : filteredSignals.length === 0 ? (
                                        <tr>
                                            <td colSpan={6} className="px-6 py-8 text-center text-gray-500">
                                                No signals found. Click "+ New Signal" to connect to another app.
                                            </td>
                                        </tr>
                                    ) : (
                                        filteredSignals.map((sig) => {
                                            const isEmitter = sig.emitter_install_id === installId;
                                            return (
                                                <tr key={sig.id} className="hover:bg-gray-50 transition-colors">
                                                    <td className="px-6 py-4 whitespace-nowrap">
                                                        <div className="flex items-center gap-2">
                                                            <Zap className="w-4 h-4 text-yellow-500 shrink-0" />
                                                            <div>
                                                                <div className="text-sm font-semibold text-gray-900">
                                                                    {sig.signal_key}
                                                                </div>
                                                                <div className="text-xs text-gray-400">
                                                                    ID #{sig.id} • Managed by: {sig.managed_by}
                                                                </div>
                                                            </div>
                                                        </div>
                                                    </td>
                                                    <td className="px-6 py-4 whitespace-nowrap">
                                                        <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                                                            isEmitter
                                                                ? 'bg-blue-100 text-blue-800'
                                                                : 'bg-gray-100 text-gray-700'
                                                        }`}>
                                                            {sig.emitter_space_id ? `Space ${sig.emitter_space_id}` : 'Package Root'}
                                                            {isEmitter && ' (This space)'}
                                                        </span>
                                                    </td>
                                                    <td className="px-6 py-4">
                                                        <div className="text-sm font-medium text-gray-900">
                                                            Space #{sig.receiver_space_id}
                                                        </div>
                                                        <div className="text-xs font-mono text-blue-600 bg-blue-50 px-1.5 py-0.5 rounded inline-block mt-0.5">
                                                            → {sig.receiver_handler}
                                                        </div>
                                                    </td>
                                                    <td className="px-6 py-4 whitespace-nowrap text-xs text-gray-500">
                                                        {sig.max_retries > 0 ? (
                                                            <div>
                                                                <span>{sig.max_retries} max retries</span>
                                                                <span className="block text-gray-400">{sig.retry_delay}s delay</span>
                                                            </div>
                                                        ) : (
                                                            <span className="text-gray-400">No retries</span>
                                                        )}
                                                    </td>
                                                    <td className="px-6 py-4 whitespace-nowrap">
                                                        <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                                                            sig.disabled
                                                                ? 'bg-red-100 text-red-800'
                                                                : 'bg-green-100 text-green-800'
                                                        }`}>
                                                            {sig.disabled ? 'Disabled' : 'Active'}
                                                        </span>
                                                    </td>
                                                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                                                        <div className="flex items-center justify-end gap-3">
                                                            <button
                                                                onClick={() => handleTestEmit(sig)}
                                                                title="Emit test event"
                                                                className="text-gray-500 hover:text-blue-600 transition-colors"
                                                            >
                                                                <Send className="w-4 h-4" />
                                                            </button>
                                                            <button
                                                                onClick={() => handleEdit(sig.id)}
                                                                title="Edit signal"
                                                                className="text-blue-600 hover:text-blue-800 transition-colors"
                                                            >
                                                                <Edit className="w-4 h-4" />
                                                            </button>
                                                            <button
                                                                onClick={() => {
                                                                    if (confirm(`Are you sure you want to delete signal "${sig.signal_key}"?`)) {
                                                                        handleDelete(sig.id);
                                                                    }
                                                                }}
                                                                title="Delete signal"
                                                                className="text-red-600 hover:text-red-800 transition-colors"
                                                            >
                                                                <Trash2 className="w-4 h-4" />
                                                            </button>
                                                        </div>
                                                    </td>
                                                </tr>
                                            );
                                        })
                                    )}
                                </tbody>
                            </table>
                        </div>
                    </div>
                ) : (
                    /* Queue / History Table */
                    <div className="bg-white rounded-lg shadow overflow-hidden border border-gray-200">
                        <div className="overflow-x-auto">
                            <table className="min-w-full divide-y divide-gray-200">
                                <thead className="bg-gray-50">
                                    <tr>
                                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                            Target / Event
                                        </th>
                                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                            Signal / Destination
                                        </th>
                                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                            Status
                                        </th>
                                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                            Retries
                                        </th>
                                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                            Error / Reason
                                        </th>
                                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                            Created At
                                        </th>
                                        <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                                            Actions
                                        </th>
                                    </tr>
                                </thead>
                                <tbody className="bg-white divide-y divide-gray-200">
                                    {queueLoader.loading ? (
                                        <tr>
                                            <td colSpan={7} className="px-6 py-6 text-center text-gray-500">
                                                Loading queue targets...
                                            </td>
                                        </tr>
                                    ) : filteredQueue.length === 0 ? (
                                        <tr>
                                            <td colSpan={7} className="px-6 py-8 text-center text-gray-500">
                                                No signal targets found in queue history.
                                            </td>
                                        </tr>
                                    ) : (
                                        filteredQueue.map((evt) => {
                                            const statusColors: Record<string, string> = {
                                                new: 'bg-yellow-100 text-yellow-800',
                                                scheduled: 'bg-blue-100 text-blue-800',
                                                delayed: 'bg-orange-100 text-orange-800',
                                                blocked: 'bg-purple-100 text-purple-800 border border-purple-200',
                                                processed: 'bg-green-100 text-green-800',
                                                failed: 'bg-red-100 text-red-800',
                                                expired: 'bg-gray-100 text-gray-800',
                                            };

                                            return (
                                                <tr key={evt.id} className="hover:bg-gray-50 transition-colors">
                                                    <td className="px-6 py-4 whitespace-nowrap text-sm">
                                                        <span className="font-mono font-medium text-gray-900">#{evt.id}</span>
                                                        <span className="block text-xs font-mono text-gray-400">Event #{evt.signal_event_id}</span>
                                                    </td>
                                                    <td className="px-6 py-4 whitespace-nowrap text-sm">
                                                        <span className="font-medium text-gray-900">{evt.signal_key || `Signal #${evt.signal_id}`}</span>
                                                        {evt.receiver_handler && (
                                                            <span className="block text-xs text-gray-500">
                                                                Space #{evt.receiver_space_id} &rarr; {evt.receiver_handler}
                                                            </span>
                                                        )}
                                                    </td>
                                                    <td className="px-6 py-4 whitespace-nowrap">
                                                        <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold uppercase ${
                                                            statusColors[evt.status] || 'bg-gray-100 text-gray-800'
                                                        }`}>
                                                            {evt.status}
                                                        </span>
                                                    </td>
                                                    <td className="px-6 py-4 whitespace-nowrap text-xs text-gray-500">
                                                        {evt.retry_count} retries
                                                    </td>
                                                    <td className="px-6 py-4 text-xs text-gray-600 max-w-xs truncate" title={evt.error}>
                                                        {evt.error || '-'}
                                                    </td>
                                                    <td className="px-6 py-4 whitespace-nowrap text-xs text-gray-500">
                                                        {evt.created_at ? new Date(evt.created_at).toLocaleString() : '-'}
                                                    </td>
                                                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm">
                                                        <div className="flex items-center justify-end gap-2">
                                                            {evt.status === 'blocked' && (
                                                                <button
                                                                    onClick={() => handleUnblock(evt.id)}
                                                                    className="px-2 py-1 bg-purple-50 hover:bg-purple-100 text-purple-700 rounded text-xs font-medium border border-purple-200 transition-colors"
                                                                    title="Unblock target to resume processing"
                                                                >
                                                                    Unblock
                                                                </button>
                                                            )}
                                                            {evt.status === 'delayed' && (
                                                                <button
                                                                    onClick={() => handleUnblock(evt.id)}
                                                                    className="px-2 py-1 bg-orange-50 hover:bg-orange-100 text-orange-700 rounded text-xs font-medium border border-orange-200 transition-colors"
                                                                    title="Retry target immediately"
                                                                >
                                                                    Retry Now
                                                                </button>
                                                            )}
                                                            <button
                                                                onClick={() => {
                                                                    if (!evt.payload) {
                                                                        setSelectedPayload("Payload was automatically purged after all targets finished processing.");
                                                                        return;
                                                                    }
                                                                    try {
                                                                        const decoded = atob(evt.payload);
                                                                        setSelectedPayload(decoded);
                                                                    } catch (e) {
                                                                        setSelectedPayload(evt.payload || '{}');
                                                                    }
                                                                }}
                                                                className="text-blue-600 hover:text-blue-800 inline-flex items-center gap-1 text-xs font-medium"
                                                            >
                                                                <Eye className="w-3.5 h-3.5" /> Payload
                                                            </button>
                                                        </div>
                                                    </td>
                                                </tr>
                                            );
                                        })
                                    )}
                                </tbody>
                            </table>
                        </div>
                    </div>
                )}

                {/* Payload Modal */}
                {selectedPayload !== null && (
                    <div className="fixed inset-0 z-50 bg-black/50 flex items-center justify-center p-4">
                        <div className="bg-white rounded-xl shadow-xl max-w-xl w-full p-6 space-y-4">
                            <h3 className="text-lg font-semibold text-gray-900">Event Payload</h3>
                            <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-xs overflow-auto max-h-96 font-mono whitespace-pre-wrap">
                                {(() => {
                                    try {
                                        return JSON.stringify(JSON.parse(selectedPayload), null, 2);
                                    } catch (e) {
                                        return selectedPayload;
                                    }
                                })()}
                            </pre>
                            <div className="flex justify-end">
                                <button
                                    onClick={() => setSelectedPayload(null)}
                                    className="px-4 py-2 bg-gray-100 hover:bg-gray-200 text-gray-700 rounded-lg text-sm font-medium transition-colors"
                                >
                                    Close
                                </button>
                            </div>
                        </div>
                    </div>
                )}
            </div>
        </WithAdminBodyLayout>
    );
};
