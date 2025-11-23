"use client";
import React, { useState, useEffect } from 'react';
import { Zap, Pencil, Target, Clock } from 'lucide-react';
import WithAdminBodyLayout from '@/contain/Layouts/WithAdminBodyLayout';
import { EventSubscription } from '@/lib';
import RuleEditor, { Rule } from './RuleEditor';

interface TargetState {
    type: string;
    endpoint: string;
    code: string; // For script type
    smtpHost: string; // For email type
    smtpPort: string;
    smtpUser: string;
    smtpPassword: string;
    smtpFrom: string;
    smtpTo: string;
    targetSpaceId: number;
}



/*

Rule Example:
    Variable      Operator       Value     ParentId
1.  orderamount   greater_than   100            
2.  $logical      group          AND                
3.  deliveryFee   less_than      10         2    
4.  paymode       equal_to       ONLINE     2    


*/


interface EventSubscriptionEditorProps {
    onSave: (data: any) => Promise<void>;
    onBack: () => void;
    initialData: EventSubscription | null;
}

export default function EventSubscriptionEditor({ onSave, onBack, initialData }: EventSubscriptionEditorProps) {
    const [eventKey, setEventKey] = useState(initialData?.event_key || '');
    const [rules, setRules] = useState<Rule[]>([]);
    const [target, setTarget] = useState<TargetState>({
        type: 'webhook',
        endpoint: '',
        code: '',
        smtpHost: '',
        smtpPort: '587',
        smtpUser: '',
        smtpPassword: '',
        smtpFrom: '',
        smtpTo: '',
        targetSpaceId: 0,
    });
    const [disabled, setDisabled] = useState(initialData?.disabled || false);
    const [delayStart, setDelayStart] = useState(initialData?.delay_start || 0);
    const [retryDelay, setRetryDelay] = useState(initialData?.retry_delay || 0);
    const [maxRetries, setMaxRetries] = useState(initialData?.max_retries || 0);
    const [saving, setSaving] = useState(false);

    useEffect(() => {
        if (initialData) {
            // Parse rules from JSON string and convert to flat array structure
            try {
                const rulesData = initialData.rules ? JSON.parse(initialData.rules) : [] as Rule[];

                // If rulesData is already a flat array (new format)
                if (Array.isArray(rulesData) && rulesData.length > 0 && rulesData[0].id) {
                    setRules(rulesData.map((r: any) => ({
                        id: r.id || `rule-${Date.now()}`,
                        variable: r.variable || '',
                        operator: r.operator || '',
                        value: r.value || '',
                        parentId: r.parent_id || r.parentId || undefined,
                    })));
                }
            } catch (e) {
                console.error('Failed to parse rules:', e);
            }

            // Parse target from subscription
            let targetData: TargetState = {
                type: initialData.target_type || 'webhook',
                endpoint: initialData.target_endpoint || '',
                code: initialData.target_code || '',
                smtpHost: '',
                smtpPort: '587',
                smtpUser: '',
                smtpPassword: '',
                smtpFrom: '',
                smtpTo: '',
                targetSpaceId: 0,
            };

            // Parse target_options for SMTP credentials if email type
            if (initialData.target_type === 'email' && initialData.target_options) {
                try {
                    const options = typeof initialData.target_options === 'string'
                        ? JSON.parse(initialData.target_options)
                        : initialData.target_options;
                    targetData.smtpHost = options.smtp_host || '';
                    targetData.smtpPort = options.smtp_port || '587';
                    targetData.smtpUser = options.smtp_user || '';
                    targetData.smtpPassword = options.smtp_password || '';
                    targetData.smtpFrom = options.smtp_from || '';
                    targetData.smtpTo = options.smtp_to || '';
                } catch (e) {
                    console.error('Failed to parse target_options:', e);
                }
            }

            if (initialData.target_type === 'space_method') {
                targetData.targetSpaceId = initialData.target_space_id || 0;
                targetData.endpoint = initialData.target_endpoint || '';
            }


            setTarget(targetData);

            // Set retry and delay fields
            setDelayStart(initialData.delay_start || 0);
            setRetryDelay(initialData.retry_delay || 0);
            setMaxRetries(initialData.max_retries || 0);
        }
    }, [initialData]);

    const handleSave = async () => {
        if (!eventKey.trim()) {
            alert('Event Key is required');
            return;
        }

        if (!target.type) {
            alert('Target type is required');
            return;
        }

        if (target.type === 'webhook' && !target.endpoint.trim()) {
            alert('Endpoint is required for webhook');
            return;
        }

        if (target.type === 'space_method' && !target.targetSpaceId) {
            alert('Space is required for space method');
            return;
        }

        if (target.type === 'space_method' && !target.endpoint.trim()) {
            alert('Event name is required for space method');
            return;
        }

        if (target.type === 'script' && !target.code.trim()) {
            alert('Code is required for script');
            return;
        }

        if (target.type === 'email') {
            if (!target.smtpHost.trim()) {
                alert('SMTP Host is required');
                return;
            }
            if (!target.smtpUser.trim()) {
                alert('SMTP User is required');
                return;
            }
            if (!target.smtpFrom.trim()) {
                alert('SMTP From is required');
                return;
            }
            if (!target.smtpTo.trim()) {
                alert('SMTP To is required');
                return;
            }
        }

        setSaving(true);
        try {

            // Build target_options based on type
            let targetOptions: any = {};
            if (target.type === 'email') {
                targetOptions = {
                    smtp_host: target.smtpHost,
                    smtp_port: target.smtpPort || '587',
                    smtp_user: target.smtpUser,
                    smtp_password: target.smtpPassword,
                    smtp_from: target.smtpFrom,
                    smtp_to: target.smtpTo,
                };
            }

            const data = {
                event_key: eventKey,
                target_type: target.type,
                target_endpoint: target.endpoint,
                target_code: target.type === 'script' ? target.code : '',
                target_options: targetOptions,
                rules: JSON.stringify(rules),
                transform: '{}',
                delay_start: delayStart,
                retry_delay: retryDelay,
                max_retries: maxRetries,
                disabled: disabled,
                target_space_id: target.targetSpaceId,
            };

            await onSave(data);
        } catch (error) {
            // Error handling is done in parent
        } finally {
            setSaving(false);
        }
    };

    const updateTarget = (updates: Partial<TargetState>) => {
        setTarget({ ...target, ...updates });
    };

    const targetTypes = [
        { value: 'email', label: 'Email' },
        { value: 'webhook', label: 'Webhook' },
        { value: 'script', label: 'Script' },
        { value: 'log', label: 'Log' },
        { value: 'space_method', label: 'Space Method' },
    ];

    return (
        <WithAdminBodyLayout
            Icon={Zap}
            name="Event Subscription"
            description={initialData ? 'Edit event subscription' : 'Create new event subscription'}
        >
            <div className="max-w-4xl mx-auto px-6 py-8 w-full space-y-6">
                {/* Event Name Section */}
                <div className="bg-white rounded-lg shadow p-6">
                    <div className="flex items-center gap-2 mb-4">
                        <Pencil className="w-5 h-5 text-gray-500" />
                        <h3 className="text-lg font-semibold text-gray-900">Event Name</h3>
                    </div>
                    <input
                        type="text"
                        value={eventKey}
                        onChange={(e) => setEventKey(e.target.value)}
                        placeholder="e.g., Record edited"
                        className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                    {eventKey && (
                        <p className="mt-2 text-sm text-gray-500">{eventKey}</p>
                    )}
                </div>

                {/* Rules Section */}
                <RuleEditor rules={rules} onRulesChange={setRules} />

                {/* Target Section */}
                <div className="bg-white rounded-lg shadow p-6">
                    <div className="flex items-center gap-2 mb-4">
                        <Target className="w-5 h-5 text-gray-500" />
                        <h3 className="text-lg font-semibold text-gray-900">Target</h3>
                    </div>

                    <div className="rounded-lg p-4">
                        <div className="mb-3">
                            <span className="text-sm font-medium text-gray-700">When rule matches</span>
                        </div>
                        <div className="space-y-3">
                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-1">Target Type</label>
                                <select
                                    value={target.type}
                                    onChange={(e) => updateTarget({ type: e.target.value })}
                                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                                >
                                    {targetTypes.map(tt => (
                                        <option key={tt.value} value={tt.value}>{tt.label}</option>
                                    ))}
                                </select>
                            </div>

                            {/* Webhook Endpoint */}
                            {target.type === 'webhook' && (
                                <div>
                                    <label className="block text-sm font-medium text-gray-700 mb-1">Endpoint URL</label>
                                    <input
                                        type="text"
                                        value={target.endpoint}
                                        onChange={(e) => updateTarget({ endpoint: e.target.value })}
                                        placeholder="https://webhook.site/..."
                                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                                    />
                                </div>
                            )}

                            {/* Script Code */}
                            {target.type === 'script' && (
                                <div>
                                    <label className="block text-sm font-medium text-gray-700 mb-1">Script Code</label>
                                    <textarea
                                        value={target.code}
                                        onChange={(e) => updateTarget({ code: e.target.value })}
                                        placeholder="// Your script code here..."
                                        rows={10}
                                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 font-mono text-sm"
                                    />
                                </div>
                            )}

                            {/* Email SMTP Credentials */}
                            {target.type === 'email' && (
                                <div className="space-y-3">
                                    <div className="grid grid-cols-2 gap-3">
                                        <div>
                                            <label className="block text-sm font-medium text-gray-700 mb-1">SMTP Host</label>
                                            <input
                                                type="text"
                                                value={target.smtpHost}
                                                onChange={(e) => updateTarget({ smtpHost: e.target.value })}
                                                placeholder="smtp.gmail.com"
                                                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                                            />
                                        </div>
                                        <div>
                                            <label className="block text-sm font-medium text-gray-700 mb-1">SMTP Port</label>
                                            <input
                                                type="text"
                                                value={target.smtpPort}
                                                onChange={(e) => updateTarget({ smtpPort: e.target.value })}
                                                placeholder="587"
                                                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                                            />
                                        </div>
                                    </div>
                                    <div className="grid grid-cols-2 gap-3">
                                        <div>
                                            <label className="block text-sm font-medium text-gray-700 mb-1">SMTP User</label>
                                            <input
                                                type="text"
                                                value={target.smtpUser}
                                                onChange={(e) => updateTarget({ smtpUser: e.target.value })}
                                                placeholder="your-email@gmail.com"
                                                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                                            />
                                        </div>
                                        <div>
                                            <label className="block text-sm font-medium text-gray-700 mb-1">SMTP Password</label>
                                            <input
                                                type="password"
                                                value={target.smtpPassword}
                                                onChange={(e) => updateTarget({ smtpPassword: e.target.value })}
                                                placeholder="your-password"
                                                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                                            />
                                        </div>
                                    </div>
                                    <div className="grid grid-cols-2 gap-3">
                                        <div>
                                            <label className="block text-sm font-medium text-gray-700 mb-1">From Email</label>
                                            <input
                                                type="email"
                                                value={target.smtpFrom}
                                                onChange={(e) => updateTarget({ smtpFrom: e.target.value })}
                                                placeholder="sender@example.com"
                                                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                                            />
                                        </div>
                                        <div>
                                            <label className="block text-sm font-medium text-gray-700 mb-1">To Email</label>
                                            <input
                                                type="email"
                                                value={target.smtpTo}
                                                onChange={(e) => updateTarget({ smtpTo: e.target.value })}
                                                placeholder="recipient@example.com"
                                                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                                            />
                                        </div>
                                    </div>
                                </div>
                            )}

                            {target.type === 'space_method' && (<>

                                <div className='flex flex-col gap-2'>
                                    <label className="block text-sm font-medium text-gray-700 mb-1">Space</label>
                                    <div className='flex gap-2'>
                                        <input
                                            type="number"
                                            placeholder="space_id"
                                            value={target.targetSpaceId}
                                            onChange={(e) => updateTarget({ targetSpaceId: parseInt(e.target.value) })}
                                            className="w-11/12 px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                                        />
                                        <button className='w-1/12 px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 hover:bg-gray-400 text-gray-00 hover:text-white'>
                                            Pick
                                        </button>
                                    </div>

                                    <label className="block text-sm font-medium text-gray-700 mb-1">Event Name</label>
                                    <input
                                        type="text"
                                        value={target.endpoint}
                                        onChange={(e) => updateTarget({ endpoint: e.target.value })}
                                        placeholder="notify_x_changed"
                                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                                    />

                                </div>


                            </>)}


                        </div>
                    </div>
                </div>

                {/* Retry and Delay Settings */}
                <div className="bg-white rounded-lg shadow p-6">
                    <div className="flex items-center gap-2 mb-4">
                        <Clock className="w-5 h-5 text-gray-500" />
                        <h3 className="text-lg font-semibold text-gray-900">Retry & Delay Settings</h3>
                    </div>
                    <div className="grid grid-cols-3 gap-4">
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">Delay Start (seconds)</label>
                            <input
                                type="number"
                                value={delayStart}
                                onChange={(e) => setDelayStart(parseInt(e.target.value) || 0)}
                                min="0"
                                placeholder="0"
                                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                            />
                            <p className="mt-1 text-xs text-gray-500">Delay before processing starts</p>
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">Retry Delay (seconds)</label>
                            <input
                                type="number"
                                value={retryDelay}
                                onChange={(e) => setRetryDelay(parseInt(e.target.value) || 0)}
                                min="0"
                                placeholder="0"
                                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                            />
                            <p className="mt-1 text-xs text-gray-500">Delay between retry attempts</p>
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">Max Retries</label>
                            <input
                                type="number"
                                value={maxRetries}
                                onChange={(e) => setMaxRetries(parseInt(e.target.value) || 0)}
                                min="0"
                                placeholder="0"
                                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                            />
                            <p className="mt-1 text-xs text-gray-500">Maximum number of retry attempts</p>
                        </div>
                    </div>
                </div>

                {/* Global Active Status */}
                <div className="bg-white rounded-lg shadow p-6">
                    <div className="flex items-center gap-2">
                        <input
                            type="checkbox"
                            checked={!disabled}
                            onChange={(e) => setDisabled(!e.target.checked)}
                            className="w-4 h-4 text-blue-600 rounded focus:ring-blue-500"
                        />
                        <label className="text-sm font-medium text-gray-700">Active</label>
                    </div>
                </div>

                {/* Footer Buttons */}
                <div className="flex justify-end gap-3">
                    <button
                        onClick={onBack}
                        className="px-6 py-2 bg-gray-200 text-gray-700 rounded-lg hover:bg-gray-300"
                    >
                        Back
                    </button>
                    <button
                        onClick={handleSave}
                        disabled={saving}
                        className="px-6 py-2 bg-teal-500 text-white rounded-lg hover:bg-teal-600 disabled:opacity-50"
                    >
                        {saving ? 'Saving...' : 'Save'}
                    </button>
                </div>
            </div>
        </WithAdminBodyLayout>
    );
}

