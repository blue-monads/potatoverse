"use client";
import React, { useState, useEffect } from 'react';
import { Zap, Pencil, Filter, Target, Plus, Trash2, Layers } from 'lucide-react';
import WithAdminBodyLayout from '@/contain/Layouts/WithAdminBodyLayout';
import { EventSubscription } from '@/lib';

interface Rule {
    id: string;
    variable: string;
    operator: string;
    value: string;
}

interface LogicalGroup {
    id: string;
    type: 'AND' | 'OR';
    rules: Rule[];
}

interface Target {
    type: string;
    endpoint: string;
    code: string; // For script type
    smtpHost: string; // For email type
    smtpPort: string;
    smtpUser: string;
    smtpPassword: string;
    smtpFrom: string;
    smtpTo: string;
}

interface EventSubscriptionEditorProps {
    onSave: (data: any) => Promise<void>;
    onBack: () => void;
    initialData: EventSubscription | null;
}

export default function EventSubscriptionEditor({ onSave, onBack, initialData }: EventSubscriptionEditorProps) {
    const [eventKey, setEventKey] = useState(initialData?.event_key || '');
    const [rules, setRules] = useState<Rule[]>([]);
    const [logicalGroups, setLogicalGroups] = useState<LogicalGroup[]>([]);
    const [target, setTarget] = useState<Target>({
        type: 'webhook',
        endpoint: '',
        code: '',
        smtpHost: '',
        smtpPort: '587',
        smtpUser: '',
        smtpPassword: '',
        smtpFrom: '',
        smtpTo: '',
    });
    const [disabled, setDisabled] = useState(initialData?.disabled || false);
    const [saving, setSaving] = useState(false);

    useEffect(() => {
        if (initialData) {
            // Parse rules from JSON string
            try {
                const rulesData = initialData.rules ? JSON.parse(initialData.rules) : {};
                if (rulesData.rules && Array.isArray(rulesData.rules)) {
                    setRules(rulesData.rules.map((r: any, idx: number) => ({
                        id: r.id || `rule-${idx}`,
                        variable: r.variable || '',
                        operator: r.operator || '',
                        value: r.value || '',
                    })));
                }
                if (rulesData.groups && Array.isArray(rulesData.groups)) {
                    setLogicalGroups(rulesData.groups.map((g: any, idx: number) => ({
                        id: g.id || `group-${idx}`,
                        type: g.type || 'AND',
                        rules: (g.rules || []).map((r: any, rIdx: number) => ({
                            id: r.id || `rule-${idx}-${rIdx}`,
                            variable: r.variable || '',
                            operator: r.operator || '',
                            value: r.value || '',
                        })),
                    })));
                }
            } catch (e) {
                console.error('Failed to parse rules:', e);
            }

            // Parse target from subscription
            let targetData: Target = {
                type: initialData.target_type || 'webhook',
                endpoint: initialData.target_endpoint || '',
                code: initialData.target_code || '',
                smtpHost: '',
                smtpPort: '587',
                smtpUser: '',
                smtpPassword: '',
                smtpFrom: '',
                smtpTo: '',
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

            setTarget(targetData);
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
            // Build rules JSON
            const rulesData = {
                rules: rules,
                groups: logicalGroups,
            };

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
                target_endpoint: target.type === 'webhook' ? target.endpoint : '',
                target_code: target.type === 'script' ? target.code : '',
                target_options: targetOptions,
                rules: JSON.stringify(rulesData),
                transform: '{}',
                disabled: disabled,
            };

            await onSave(data);
        } catch (error) {
            // Error handling is done in parent
        } finally {
            setSaving(false);
        }
    };

    const addRule = () => {
        setRules([...rules, {
            id: `rule-${Date.now()}`,
            variable: '',
            operator: 'equal_to',
            value: '',
        }]);
    };

    const updateRule = (id: string, updates: Partial<Rule>) => {
        setRules(rules.map(r => r.id === id ? { ...r, ...updates } : r));
    };

    const deleteRule = (id: string) => {
        setRules(rules.filter(r => r.id !== id));
    };

    const addLogicalGroup = () => {
        setLogicalGroups([...logicalGroups, {
            id: `group-${Date.now()}`,
            type: 'AND',
            rules: [{
                id: `rule-${Date.now()}-1`,
                variable: '',
                operator: 'equal_to',
                value: '',
            }],
        }]);
    };

    const addRuleToGroup = (groupId: string) => {
        setLogicalGroups(logicalGroups.map(g => 
            g.id === groupId 
                ? { ...g, rules: [...g.rules, { id: `rule-${Date.now()}`, variable: '', operator: 'equal_to', value: '' }] }
                : g
        ));
    };

    const updateRuleInGroup = (groupId: string, ruleId: string, updates: Partial<Rule>) => {
        setLogicalGroups(logicalGroups.map(g =>
            g.id === groupId
                ? { ...g, rules: g.rules.map(r => r.id === ruleId ? { ...r, ...updates } : r) }
                : g
        ));
    };

    const deleteRuleFromGroup = (groupId: string, ruleId: string) => {
        setLogicalGroups(logicalGroups.map(g =>
            g.id === groupId
                ? { ...g, rules: g.rules.filter(r => r.id !== ruleId) }
                : g
        ));
    };

    const toggleGroupType = (groupId: string) => {
        setLogicalGroups(logicalGroups.map(g =>
            g.id === groupId
                ? { ...g, type: g.type === 'AND' ? 'OR' : 'AND' }
                : g
        ));
    };

    const deleteGroup = (groupId: string) => {
        setLogicalGroups(logicalGroups.filter(g => g.id !== groupId));
    };

    const updateTarget = (updates: Partial<Target>) => {
        setTarget({ ...target, ...updates });
    };

    const operators = [
        { value: 'equal_to', label: 'Equal To' },
        { value: 'not_equal_to', label: 'Not Equal To' },
        { value: 'greater_than', label: 'Greater Than' },
        { value: 'less_than', label: 'Less Than' },
        { value: 'greater_than_or_equal', label: 'Greater Than Or Equal' },
        { value: 'less_than_or_equal', label: 'Less Than Or Equal' },
        { value: 'contains', label: 'Contains' },
        { value: 'not_contains', label: 'Not Contains' },
        { value: 'before', label: 'Before' },
        { value: 'after', label: 'After' },
    ];

    const targetTypes = [
        { value: 'webhook', label: 'Webhook' },
        { value: 'email', label: 'Email' },
        { value: 'script', label: 'Script' },
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
                <div className="bg-white rounded-lg shadow p-6">
                    <div className="flex items-center gap-2 mb-4">
                        <Filter className="w-5 h-5 text-gray-500" />
                        <h3 className="text-lg font-semibold text-gray-900">Rules</h3>
                    </div>

                    {/* Individual Rules */}
                    <div className="space-y-3 mb-4">
                        {rules.map((rule) => (
                            <div key={rule.id} className="flex items-center gap-3 p-3 border border-gray-200 rounded-lg">
                                <input
                                    type="text"
                                    value={rule.variable}
                                    onChange={(e) => updateRule(rule.id, { variable: e.target.value })}
                                    placeholder="Variable"
                                    className="flex-1 px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
                                />
                                <select
                                    value={rule.operator}
                                    onChange={(e) => updateRule(rule.id, { operator: e.target.value })}
                                    className="px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
                                >
                                    {operators.map(op => (
                                        <option key={op.value} value={op.value}>{op.label}</option>
                                    ))}
                                </select>
                                <input
                                    type="text"
                                    value={rule.value}
                                    onChange={(e) => updateRule(rule.id, { value: e.target.value })}
                                    placeholder="Value"
                                    className="flex-1 px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
                                />
                                <button
                                    onClick={() => deleteRule(rule.id)}
                                    className="text-red-600 hover:text-red-900"
                                >
                                    <Trash2 className="w-4 h-4" />
                                </button>
                            </div>
                        ))}
                    </div>

                    <div className="flex gap-2 mb-4">
                        <button
                            onClick={addRule}
                            className="flex items-center gap-2 px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600"
                        >
                            <Plus className="w-4 h-4" />
                            Add Rule
                        </button>
                        <button
                            onClick={addLogicalGroup}
                            className="flex items-center gap-2 px-4 py-2 bg-green-500 text-white rounded-lg hover:bg-green-600"
                        >
                            <Plus className="w-4 h-4" />
                            Logical Group (AND)
                        </button>
                        <button
                            onClick={addLogicalGroup}
                            className="flex items-center gap-2 px-4 py-2 bg-green-500 text-white rounded-lg hover:bg-green-600"
                        >
                            <Plus className="w-4 h-4" />
                            Logical Group (OR)
                        </button>
                    </div>

                    {/* Logical Groups */}
                    <div className="space-y-3">
                        {logicalGroups.map((group) => (
                            <div key={group.id} className="border border-green-300 rounded-lg p-4 bg-green-50">
                                <div className="flex items-center justify-between mb-3">
                                    <div className="flex items-center gap-2">
                                        <Layers className="w-4 h-4 text-green-700" />
                                        <span className="font-medium text-green-900">
                                            Logical Group ({group.type})
                                        </span>
                                    </div>
                                    <div className="flex items-center gap-2">
                                        <button
                                            onClick={() => toggleGroupType(group.id)}
                                            className="px-3 py-1 text-sm bg-green-600 text-white rounded hover:bg-green-700"
                                        >
                                            Switch to {group.type === 'AND' ? 'OR' : 'AND'}
                                        </button>
                                        <button
                                            onClick={() => deleteGroup(group.id)}
                                            className="text-red-600 hover:text-red-900"
                                        >
                                            <Trash2 className="w-4 h-4" />
                                        </button>
                                    </div>
                                </div>
                                <div className="space-y-2">
                                    {group.rules.map((rule) => (
                                        <div key={rule.id} className="flex items-center gap-3 p-2 bg-white rounded">
                                            <input
                                                type="text"
                                                value={rule.variable}
                                                onChange={(e) => updateRuleInGroup(group.id, rule.id, { variable: e.target.value })}
                                                placeholder="Variable"
                                                className="flex-1 px-3 py-2 border border-gray-300 rounded text-sm"
                                            />
                                            <select
                                                value={rule.operator}
                                                onChange={(e) => updateRuleInGroup(group.id, rule.id, { operator: e.target.value })}
                                                className="px-3 py-2 border border-gray-300 rounded text-sm"
                                            >
                                                {operators.map(op => (
                                                    <option key={op.value} value={op.value}>{op.label}</option>
                                                ))}
                                            </select>
                                            <input
                                                type="text"
                                                value={rule.value}
                                                onChange={(e) => updateRuleInGroup(group.id, rule.id, { value: e.target.value })}
                                                placeholder="Value"
                                                className="flex-1 px-3 py-2 border border-gray-300 rounded text-sm"
                                            />
                                            <button
                                                onClick={() => deleteRuleFromGroup(group.id, rule.id)}
                                                className="text-red-600 hover:text-red-900"
                                            >
                                                <Trash2 className="w-4 h-4" />
                                            </button>
                                        </div>
                                    ))}
                                </div>
                                <button
                                    onClick={() => addRuleToGroup(group.id)}
                                    className="mt-3 flex items-center gap-2 px-3 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 text-sm"
                                >
                                    <Plus className="w-4 h-4" />
                                    Add Rule
                                </button>
                            </div>
                        ))}
                    </div>
                </div>

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

