"use client";
import React, { useState } from 'react';
import { SpaceCapability, CapabilityDefinition } from '@/lib';
import { buildCapabilityOptions } from '@/contain/compo/capabilityUtils';
import { CapabilityOptionsSection } from '../sub/CapabilityOptionsSection';

const CapEditor = ({
    capability,
    definition,
    capabilityTypes,
    onSave,
    onCancel
}: {
    capability: SpaceCapability;
    definition?: CapabilityDefinition;
    capabilityTypes: CapabilityDefinition[];
    onSave: (data: any) => void;
    onCancel: () => void;
}) => {
    const [name, setName] = useState(capability.name);
    const [capabilityType, setCapabilityType] = useState(capability.capability_type);
    const [spaceId, setSpaceId] = useState(capability.space_id);
    const [formData, setFormData] = useState<Record<string, any>>(() => {
        try {
            return JSON.parse(capability.options || '{}');
        } catch {
            return {};
        }
    });

    const currentDefinition = capabilityTypes.find(t => t.name === capabilityType);

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        
        const options = buildCapabilityOptions(currentDefinition, formData, false);

        onSave({
            name,
            capability_type: capabilityType,
            space_id: spaceId,
            options,
        });
    };

    return (
        <form onSubmit={handleSubmit} className="space-y-4 bg-gray-50 p-4 rounded-lg">
            <div className="grid grid-cols-2 gap-4">
                <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                        Name *
                    </label>
                    <input
                        type="text"
                        value={name}
                        onChange={(e) => setName(e.target.value)}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                        required
                    />
                </div>
                <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                        Type *
                    </label>
                    <select
                        value={capabilityType}
                        onChange={(e) => {
                            setCapabilityType(e.target.value);
                            setFormData({});
                        }}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                        required
                    >
                        {capabilityTypes.map(type => (
                            <option key={type.name} value={type.name}>
                                {type.name}
                            </option>
                        ))}
                    </select>
                </div>
            </div>
            <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                    Scope
                </label>
                <select
                    value={spaceId}
                    onChange={(e) => setSpaceId(parseInt(e.target.value))}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                >
                    <option value={0}>Package Level (root)</option>
                    {spaceId > 0 && <option value={spaceId}>Space #{spaceId}</option>}
                </select>
            </div>

            {currentDefinition && (
                <CapabilityOptionsSection
                    definition={currentDefinition}
                    formData={formData}
                    onFieldChange={(key, value) => setFormData({ ...formData, [key]: value })}
                    className="bg-white"
                />
            )}

            <div className="flex justify-end gap-2">
                <button
                    type="button"
                    onClick={onCancel}
                    className="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50"
                >
                    Cancel
                </button>
                <button
                    type="submit"
                    className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600"
                >
                    Save
                </button>
            </div>
        </form>
    );
};

export default CapEditor;