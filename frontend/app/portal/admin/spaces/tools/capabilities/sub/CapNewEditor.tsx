import { CapabilityDefinition } from '@/lib';
import { CapabilityOptionsSection } from '../sub/CapabilityOptionsSection';
import { buildCapabilityOptions } from '@/contain/compo/capabilityUtils';
import { useState } from 'react';


const CapNewEditor = ({
    capabilityType,
    capabilityTypes,
    defaultSpaceId,
    onSave,
    onCancel
}: {
    capabilityType: string;
    capabilityTypes: CapabilityDefinition[];
    defaultSpaceId: number;
    onSave: (data: any) => void;
    onCancel: () => void;
}) => {
    const [selectedType, setSelectedType] = useState(capabilityType);
    const [name, setName] = useState('');
    const [spaceId, setSpaceId] = useState(defaultSpaceId);
    const [formData, setFormData] = useState<Record<string, any>>({});
    const [saving, setSaving] = useState(false);

    const definition = capabilityTypes.find(t => t.name === selectedType);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!name || !selectedType) {
            alert('Name and type are required');
            return;
        }

        setSaving(true);
        try {
            const options = buildCapabilityOptions(definition, formData, true);

            await onSave({
                name,
                capability_type: selectedType,
                space_id: spaceId,
                options,
            });
        } catch (error) {
            // Error already handled in parent
        } finally {
            setSaving(false);
        }
    };

    return (
        <div className="bg-white rounded-lg shadow p-6">
            <h2 className="text-xl font-bold mb-6">Create Capability</h2>
            <form onSubmit={handleSubmit} className="space-y-4">
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
                        Capability Type *
                    </label>
                    <select
                        value={selectedType}
                        onChange={(e) => {
                            setSelectedType(e.target.value);
                            setFormData({});
                        }}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                        required
                    >
                        <option value="">Select a type...</option>
                        {capabilityTypes.map(type => (
                            <option key={type.name} value={type.name}>
                                {type.name}
                            </option>
                        ))}
                    </select>
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
                        {defaultSpaceId > 0 && <option value={defaultSpaceId}>This Space (#{defaultSpaceId})</option>}
                    </select>
                    <p className="text-xs text-gray-500 mt-1">
                        Package level applies to all spaces. Space level applies only to a specific space.
                    </p>
                </div>

                {definition && (
                    <CapabilityOptionsSection
                        definition={definition}
                        formData={formData}
                        onFieldChange={(key, value) => setFormData({ ...formData, [key]: value })}
                    />
                )}

                <div className="flex justify-end gap-2 pt-4">
                    <button
                        type="button"
                        onClick={onCancel}
                        disabled={saving}
                        className="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 disabled:opacity-50"
                    >
                        Cancel
                    </button>
                    <button
                        type="submit"
                        disabled={saving}
                        className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50"
                    >
                        {saving ? 'Creating...' : 'Create'}
                    </button>
                </div>
            </form>
        </div>
    );
};

export default CapNewEditor;
