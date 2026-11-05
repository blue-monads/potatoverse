"use client";
import React, { useState, useEffect } from 'react';
import { Zap, Target, Clock, ArrowLeft, List, Shield, RefreshCw } from 'lucide-react';
import WithAdminBodyLayout from '@/contain/Layouts/WithAdminBodyLayout';
import { Signal, getSpaceSpec, listInstalledSpaces, Space } from '@/lib';

interface SpaceSpec {
    space_specs: Record<string, {
        events_outputs?: Array<{
            name: string;
            description: string;
            schema?: any;
        }>;
        event_slots?: Array<{
            name: string;
            description: string;
            schema?: any;
        }>;
    }>;
}

interface SignalEditorProps {
    onSave: (data: Partial<Signal>) => Promise<void>;
    onBack: () => void;
    initialData: Signal | null;
    installId: number;
    spaceId?: number;
}

export default function SignalEditor({ onSave, onBack, initialData, installId, spaceId }: SignalEditorProps) {
    const [signalKey, setSignalKey] = useState(initialData?.signal_key || '');
    const [receiverSpaceId, setReceiverSpaceId] = useState<number>(initialData?.receiver_space_id || 0);
    const [receiverInstallId, setReceiverInstallId] = useState<number>(initialData?.receiver_install_id || 0);
    const [receiverHandler, setReceiverHandler] = useState(initialData?.receiver_handler || '');
    const [managedBy, setManagedBy] = useState<'both' | 'emitter' | 'receiver'>(initialData?.managed_by || 'both');
    const [maxRetries, setMaxRetries] = useState(initialData?.max_retries ?? 3);
    const [retryDelay, setRetryDelay] = useState(initialData?.retry_delay ?? 5);
    const [expiresOn, setExpiresOn] = useState<number>(initialData?.expires_on || 0);
    const [disabled, setDisabled] = useState(initialData?.disabled || false);
    const [saving, setSaving] = useState(false);

    // Emitter events dropdown
    const [showEventsDropdown, setShowEventsDropdown] = useState(false);
    const [availableEvents, setAvailableEvents] = useState<Array<{ name: string; description: string }>>([]);
    const [loadingEvents, setLoadingEvents] = useState(false);

    // Spaces dropdown
    const [availableSpaces, setAvailableSpaces] = useState<Space[]>([]);
    const [loadingSpaces, setLoadingSpaces] = useState(false);

    // Receiver event slots dropdown
    const [showSlotsDropdown, setShowSlotsDropdown] = useState(false);
    const [availableSlots, setAvailableSlots] = useState<Array<{ name: string; description: string }>>([]);
    const [loadingSlots, setLoadingSlots] = useState(false);

    useEffect(() => {
        if (installId) {
            fetchAvailableEvents();
            fetchAvailableSpaces();
        }
    }, [installId]);

    useEffect(() => {
        if (receiverSpaceId && availableSpaces.length > 0) {
            const targetSpace = availableSpaces.find(s => s.id === receiverSpaceId);
            if (targetSpace) {
                setReceiverInstallId(targetSpace.install_id);
                fetchSlotsForSpace(targetSpace.install_id);
            }
        }
    }, [receiverSpaceId, availableSpaces]);

    const fetchAvailableEvents = async () => {
        setLoadingEvents(true);
        try {
            const response = await getSpaceSpec(installId);
            const specData: SpaceSpec = response.data;
            const events: Array<{ name: string; description: string }> = [];
            Object.values(specData.space_specs || {}).forEach(spaceSpec => {
                if (spaceSpec.events_outputs) {
                    spaceSpec.events_outputs.forEach(event => {
                        events.push({
                            name: event.name,
                            description: event.description,
                        });
                    });
                }
            });
            setAvailableEvents(events);
        } catch (error) {
            console.error('Error fetching available events:', error);
        } finally {
            setLoadingEvents(false);
        }
    };

    const fetchAvailableSpaces = async () => {
        setLoadingSpaces(true);
        try {
            const response = await listInstalledSpaces();
            const spaces = response.data.spaces || [];
            setAvailableSpaces(spaces);
        } catch (error) {
            console.error('Error fetching available spaces:', error);
        } finally {
            setLoadingSpaces(false);
        }
    };

    const fetchSlotsForSpace = async (targetInstallId: number) => {
        setLoadingSlots(true);
        try {
            const response = await getSpaceSpec(targetInstallId);
            const specData: SpaceSpec = response.data;
            const slots: Array<{ name: string; description: string }> = [];
            Object.values(specData.space_specs || {}).forEach(spaceSpec => {
                if (spaceSpec.event_slots) {
                    spaceSpec.event_slots.forEach(slot => {
                        slots.push({
                            name: slot.name,
                            description: slot.description,
                        });
                    });
                }
            });
            setAvailableSlots(slots);
        } catch (error) {
            console.error('Error fetching slots for space:', error);
            setAvailableSlots([]);
        } finally {
            setLoadingSlots(false);
        }
    };

    const handleSave = async () => {
        if (!signalKey.trim()) {
            alert('Signal key is required');
            return;
        }

        if (!receiverSpaceId) {
            alert('Receiver space is required');
            return;
        }

        if (!receiverHandler.trim()) {
            alert('Receiver handler is required');
            return;
        }

        setSaving(true);
        try {
            const payload: Partial<Signal> = {
                signal_key: signalKey.trim(),
                emitter_install_id: installId,
                emitter_space_id: spaceId || 0,
                receiver_install_id: receiverInstallId,
                receiver_space_id: receiverSpaceId,
                receiver_handler: receiverHandler.trim(),
                managed_by: managedBy,
                max_retries: Number(maxRetries),
                retry_delay: Number(retryDelay),
                expires_on: Number(expiresOn),
                disabled: disabled,
            };

            await onSave(payload);
        } catch (error) {
            // Error handled by caller
        } finally {
            setSaving(false);
        }
    };

    return (
        <WithAdminBodyLayout
            Icon={Zap}
            name={initialData ? "Edit Signal" : "New Signal"}
            description="Connect an emitted signal to a receiver app handler"
            variant="none"
        >
            <div className="max-w-4xl mx-auto px-4 py-6 flex flex-col gap-6">
                <button
                    onClick={onBack}
                    className="inline-flex items-center gap-2 text-gray-600 hover:text-gray-900 transition-colors w-fit"
                >
                    <ArrowLeft className="w-4 h-4" />
                    <span className="text-sm font-semibold">Back to Signals</span>
                </button>

                {/* Signal Key Section */}
                <div className="bg-white rounded-lg shadow p-6 border border-gray-100">
                    <div className="flex items-center gap-2 mb-4">
                        <Zap className="w-5 h-5 text-yellow-500" />
                        <h3 className="text-lg font-semibold text-gray-900">Emitted Signal</h3>
                    </div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                        Signal Key (Event Name)
                    </label>
                    <div className="relative">
                        <div className="flex gap-2">
                            <input
                                type="text"
                                value={signalKey}
                                onChange={(e) => setSignalKey(e.target.value)}
                                placeholder="e.g. order_created or table.row_updated"
                                className="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                            />
                            <button
                                type="button"
                                onClick={() => setShowEventsDropdown(!showEventsDropdown)}
                                disabled={loadingEvents}
                                className="px-4 py-2 bg-gray-100 border border-gray-300 rounded-lg hover:bg-gray-200 focus:outline-none focus:ring-2 focus:ring-blue-500 flex items-center gap-2 text-sm"
                            >
                                <List className="w-4 h-4" />
                                {loadingEvents ? 'Loading...' : 'Known Events'}
                            </button>
                        </div>

                        {showEventsDropdown && availableEvents.length > 0 && (
                            <div className="absolute z-20 mt-1 w-full bg-white border border-gray-300 rounded-lg shadow-lg max-h-60 overflow-y-auto">
                                {availableEvents.map((evt, idx) => (
                                    <button
                                        key={idx}
                                        type="button"
                                        onClick={() => {
                                            setSignalKey(evt.name);
                                            setShowEventsDropdown(false);
                                        }}
                                        className="w-full px-4 py-2.5 text-left hover:bg-blue-50 border-b border-gray-100 last:border-b-0 transition-colors"
                                    >
                                        <div className="font-medium text-gray-900 text-sm">{evt.name}</div>
                                        {evt.description && (
                                            <div className="text-xs text-gray-500">{evt.description}</div>
                                        )}
                                    </button>
                                ))}
                            </div>
                        )}

                        {showEventsDropdown && availableEvents.length === 0 && !loadingEvents && (
                            <div className="absolute z-20 mt-1 w-full bg-white border border-gray-300 rounded-lg shadow-lg px-4 py-3 text-sm text-gray-500">
                                No documented event outputs found in package spec. You can enter any custom signal key.
                            </div>
                        )}
                    </div>
                </div>

                {/* Receiver Target App Section */}
                <div className="bg-white rounded-lg shadow p-6 border border-gray-100">
                    <div className="flex items-center gap-2 mb-4">
                        <Target className="w-5 h-5 text-blue-500" />
                        <h3 className="text-lg font-semibold text-gray-900">Connected App (Receiver)</h3>
                    </div>

                    <div className="space-y-4">
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">
                                Target Space
                            </label>
                            <select
                                value={receiverSpaceId}
                                onChange={(e) => {
                                    const sId = parseInt(e.target.value, 10);
                                    setReceiverSpaceId(sId);
                                    const space = availableSpaces.find(s => s.id === sId);
                                    if (space) {
                                        setReceiverInstallId(space.install_id);
                                    }
                                    setReceiverHandler('');
                                }}
                                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                            >
                                <option value={0}>Select a receiver space...</option>
                                {availableSpaces.map((s) => (
                                    <option key={s.id} value={s.id}>
                                        {s.name || `Space #${s.id}`} ({s.namespace_key}) [ID: {s.id}]
                                    </option>
                                ))}
                            </select>
                        </div>

                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">
                                Receiver Handler / Action
                            </label>
                            <div className="relative">
                                <div className="flex gap-2">
                                    <input
                                        type="text"
                                        value={receiverHandler}
                                        onChange={(e) => setReceiverHandler(e.target.value)}
                                        placeholder="e.g. handle_order or on_record_change"
                                        className="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                                    />
                                    <button
                                        type="button"
                                        onClick={() => setShowSlotsDropdown(!showSlotsDropdown)}
                                        disabled={!receiverSpaceId || loadingSlots}
                                        className="px-4 py-2 bg-gray-100 border border-gray-300 rounded-lg hover:bg-gray-200 focus:outline-none focus:ring-2 focus:ring-blue-500 flex items-center gap-2 text-sm disabled:opacity-50 disabled:cursor-not-allowed"
                                    >
                                        <List className="w-4 h-4" />
                                        {loadingSlots ? 'Loading...' : 'Known Slots'}
                                    </button>
                                </div>

                                {showSlotsDropdown && availableSlots.length > 0 && (
                                    <div className="absolute z-20 mt-1 w-full bg-white border border-gray-300 rounded-lg shadow-lg max-h-60 overflow-y-auto">
                                        {availableSlots.map((slot, idx) => (
                                            <button
                                                key={idx}
                                                type="button"
                                                onClick={() => {
                                                    setReceiverHandler(slot.name);
                                                    setShowSlotsDropdown(false);
                                                }}
                                                className="w-full px-4 py-2.5 text-left hover:bg-blue-50 border-b border-gray-100 last:border-b-0 transition-colors"
                                            >
                                                <div className="font-medium text-gray-900 text-sm">{slot.name}</div>
                                                {slot.description && (
                                                    <div className="text-xs text-gray-500">{slot.description}</div>
                                                )}
                                            </button>
                                        ))}
                                    </div>
                                )}

                                {showSlotsDropdown && availableSlots.length === 0 && !loadingSlots && (
                                    <div className="absolute z-20 mt-1 w-full bg-white border border-gray-300 rounded-lg shadow-lg px-4 py-3 text-sm text-gray-500">
                                        No documented event slots found for this space. You can enter any custom handler name.
                                    </div>
                                )}
                            </div>
                        </div>

                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">
                                Managed By
                            </label>
                            <select
                                value={managedBy}
                                onChange={(e) => setManagedBy(e.target.value as any)}
                                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                            >
                                <option value="both">Both (Emitter & Receiver spaces can edit/view)</option>
                                <option value="emitter">Emitter Only</option>
                                <option value="receiver">Receiver Only</option>
                            </select>
                        </div>
                    </div>
                </div>

                {/* Delivery & Retry Options */}
                <div className="bg-white rounded-lg shadow p-6 border border-gray-100">
                    <div className="flex items-center gap-2 mb-4">
                        <RefreshCw className="w-5 h-5 text-indigo-500" />
                        <h3 className="text-lg font-semibold text-gray-900">Delivery & Retry Policy</h3>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">
                                Max Retries
                            </label>
                            <input
                                type="number"
                                min={0}
                                max={20}
                                value={maxRetries}
                                onChange={(e) => setMaxRetries(parseInt(e.target.value, 10) || 0)}
                                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                            />
                            <p className="text-xs text-gray-500 mt-1">Number of times to retry failed dispatches (0 for no retry)</p>
                        </div>

                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">
                                Retry Delay (Seconds)
                            </label>
                            <input
                                type="number"
                                min={0}
                                max={86400}
                                value={retryDelay}
                                onChange={(e) => setRetryDelay(parseInt(e.target.value, 10) || 0)}
                                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                            />
                            <p className="text-xs text-gray-500 mt-1">Time to wait between dispatch attempts</p>
                        </div>

                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">
                                Expires On (Unix Timestamp)
                            </label>
                            <input
                                type="number"
                                min={0}
                                value={expiresOn}
                                onChange={(e) => setExpiresOn(parseInt(e.target.value, 10) || 0)}
                                placeholder="0 for never"
                                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                            />
                            <p className="text-xs text-gray-500 mt-1">Optional expiration timestamp (0 = never expires)</p>
                        </div>

                        <div className="flex items-center h-full pt-6">
                            <label className="flex items-center gap-3 cursor-pointer">
                                <input
                                    type="checkbox"
                                    checked={disabled}
                                    onChange={(e) => setDisabled(e.target.checked)}
                                    className="w-4 h-4 text-blue-600 rounded border-gray-300 focus:ring-blue-500"
                                />
                                <span className="text-sm font-medium text-gray-700">Disable Signal</span>
                            </label>
                        </div>
                    </div>
                </div>

                {/* Action Buttons */}
                <div className="flex justify-end gap-3 pt-2">
                    <button
                        type="button"
                        onClick={onBack}
                        className="px-5 py-2.5 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-50 transition-colors"
                    >
                        Cancel
                    </button>
                    <button
                        type="button"
                        onClick={handleSave}
                        disabled={saving}
                        className="px-5 py-2.5 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 transition-colors disabled:opacity-50"
                    >
                        {saving ? 'Saving...' : (initialData ? 'Update Signal' : 'Create Signal')}
                    </button>
                </div>
            </div>
        </WithAdminBodyLayout>
    );
}
