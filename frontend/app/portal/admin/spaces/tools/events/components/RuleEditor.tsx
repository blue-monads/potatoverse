"use client";
import React from 'react';
import { Filter, Plus, Trash2, Layers } from 'lucide-react';

export interface Rule {
    id: string;
    variable: string;
    operator: string;
    value: string;
    parentId?: string; // Optional parent ID for grouping
}

interface RuleEditorProps {
    rules: Rule[];
    onRulesChange: (rules: Rule[]) => void;
}

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
    { value: 'group', label: 'Logical Group' }, // Special operator for logical groups
];

export default function RuleEditor({ rules, onRulesChange }: RuleEditorProps) {
    const isLogicalGroup = (rule: Rule): boolean => {
        return rule.variable === '$logical' && rule.operator === 'group';
    };

    const getGroupType = (rule: Rule): 'AND' | 'OR' => {
        return (rule.value === 'OR' ? 'OR' : 'AND') as 'AND' | 'OR';
    };

    const getChildRules = (parentId: string): Rule[] => {
        return rules.filter(r => r.parentId === parentId);
    };

    const getTopLevelRules = (): Rule[] => {
        return rules.filter(r => !r.parentId);
    };

    const addRule = () => {
        const newRule: Rule = {
            id: `rule-${Date.now()}`,
            variable: '',
            operator: 'equal_to',
            value: '',            
        };
        onRulesChange([...rules, newRule]);
    };

    const addLogicalGroup = (type: 'AND' | 'OR') => {
        const groupId = `group-${Date.now()}`;
        const groupRule: Rule = {
            id: groupId,
            variable: '$logical',
            operator: 'group',
            value: type,
        };
        // Add a default rule inside the group
        const childRule: Rule = {
            id: `rule-${Date.now()}-1`,
            variable: '',
            operator: 'equal_to',
            value: '',
            parentId: groupId,
        };
        onRulesChange([...rules, groupRule, childRule]);
    };

    const updateRule = (id: string, updates: Partial<Rule>) => {
        onRulesChange(rules.map(r => r.id === id ? { ...r, ...updates } : r));
    };

    const deleteRule = (id: string) => {
        // Delete the rule and all its children
        const idsToDelete = new Set<string>([id]);
        const findChildren = (parentId: string) => {
            rules.forEach(r => {
                if (r.parentId === parentId) {
                    idsToDelete.add(r.id);
                    findChildren(r.id);
                }
            });
        };
        findChildren(id);
        onRulesChange(rules.filter(r => !idsToDelete.has(r.id)));
    };

    const addRuleToGroup = (groupId: string) => {
        const newRule: Rule = {
            id: `rule-${Date.now()}`,
            variable: '',
            operator: 'equal_to',
            value: '',
            parentId: groupId,
        };
        onRulesChange([...rules, newRule]);
    };

    const toggleGroupType = (groupId: string) => {
        const groupRule = rules.find(r => r.id === groupId);
        if (groupRule && isLogicalGroup(groupRule)) {
            const newType = getGroupType(groupRule) === 'AND' ? 'OR' : 'AND';
            updateRule(groupId, { value: newType });
        }
    };

    const renderRule = (rule: Rule, level: number = 0) => {
        if (isLogicalGroup(rule)) {
            const groupType = getGroupType(rule);
            const childRules = getChildRules(rule.id);
            
            return (
                <div key={rule.id} className={`border border-green-300 rounded-lg p-4 bg-green-50 ${level > 0 ? 'ml-6' : ''}`}>
                    <div className="flex items-center justify-between mb-3">
                        <div className="flex items-center gap-2">
                            <Layers className="w-4 h-4 text-green-700" />
                            <span className="font-medium text-green-900">
                                Logical Group ({groupType})
                            </span>
                        </div>
                        <div className="flex items-center gap-2">
                            <button
                                onClick={() => toggleGroupType(rule.id)}
                                className="px-3 py-1 text-sm bg-green-600 text-white rounded hover:bg-green-700"
                            >
                                Switch to {groupType === 'AND' ? 'OR' : 'AND'}
                            </button>
                            <button
                                onClick={() => deleteRule(rule.id)}
                                className="text-red-600 hover:text-red-900"
                            >
                                <Trash2 className="w-4 h-4" />
                            </button>
                        </div>
                    </div>
                    <div className="space-y-2">
                        {childRules.map(childRule => renderRule(childRule, level + 1))}
                    </div>
                    <button
                        onClick={() => addRuleToGroup(rule.id)}
                        className="mt-3 flex items-center gap-2 px-3 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 text-sm"
                    >
                        <Plus className="w-4 h-4" />
                        Add Rule to Group
                    </button>
                </div>
            );
        }

        return (
            <div key={rule.id} className={`flex items-center gap-3 p-3 border border-gray-200 rounded-lg ${level > 0 ? 'ml-6 bg-white' : ''}`}>
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
        );
    };

    return (
        <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center gap-2 mb-4">
                <Filter className="w-5 h-5 text-gray-500" />
                <h3 className="text-lg font-semibold text-gray-900">Rules</h3>
            </div>

            {/* Rules List */}
            <div className="space-y-3 mb-4">
                {getTopLevelRules().map(rule => renderRule(rule))}
            </div>

            {/* Action Buttons */}
            <div className="flex gap-2">
                <button
                    onClick={addRule}
                    className="flex items-center gap-2 px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600"
                >
                    <Plus className="w-4 h-4" />
                    Add Rule
                </button>
                <button
                    onClick={() => addLogicalGroup('AND')}
                    className="flex items-center gap-2 px-4 py-2 bg-green-500 text-white rounded-lg hover:bg-green-600"
                >
                    <Plus className="w-4 h-4" />
                    Logical Group (AND)
                </button>
                <button
                    onClick={() => addLogicalGroup('OR')}
                    className="flex items-center gap-2 px-4 py-2 bg-green-500 text-white rounded-lg hover:bg-green-600"
                >
                    <Plus className="w-4 h-4" />
                    Logical Group (OR)
                </button>
            </div>
        </div>
    );
}

