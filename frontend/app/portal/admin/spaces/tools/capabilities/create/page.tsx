"use client";
import React, { useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { ArrowLeft } from 'lucide-react';
import WithAdminBodyLayout from '@/contain/Layouts/WithAdminBodyLayout';
import { 
    createSpaceCapability,
    listCapabilityTypes,
    CapabilityDefinition
} from '@/lib';
import useSimpleDataLoader from '@/hooks/useSimpleDataLoader';
import { CapabilityOptionsSection } from '../sub/CapabilityOptionsSection';
import { buildCapabilityOptions } from '@/contain/compo/capabilityUtils';

export default function Page() {
    const router = useRouter();
    const searchParams = useSearchParams();
    const installId = searchParams.get('install_id');
    const spaceId = searchParams.get('space_id');

    if (!installId) {
        return <div>Install ID not provided</div>;
    }

    const capabilityTypesLoader = useSimpleDataLoader<CapabilityDefinition[]>({
        loader: () => listCapabilityTypes(),
        ready: true,
    });

    const handleSave = async (data: {
        name: string;
        capability_type: string;
        space_id?: number;
        options?: any;
        extrameta?: any;
    }) => {
        try {
            await createSpaceCapability(parseInt(installId), data);
            const params = new URLSearchParams();
            params.set('install_id', installId);
            if (spaceId) params.set('space_id', spaceId);
            router.push(`/portal/admin/spaces/tools/capabilities?${params.toString()}`);
        } catch (error) {
            console.error('Failed to create capability:', error);
            alert('Failed to create capability: ' + ((error as any)?.response?.data?.error || (error as any)?.message));
            throw error;
        }
    };

    const handleBack = () => {
        const params = new URLSearchParams();
        params.set('install_id', installId);
        if (spaceId) params.set('space_id', spaceId);
        router.push(`/portal/admin/spaces/tools/capabilities?${params.toString()}`);
    };

    return (
        <WithAdminBodyLayout
            Icon={ArrowLeft}
            name="Create Capability"
            description="Add a new capability to this package or space"
            rightContent={
                <button
                    onClick={handleBack}
                    className="flex items-center gap-2 px-4 py-2 text-gray-600 hover:text-gray-900"
                >
                    <ArrowLeft className="w-4 h-4" />
                    Back
                </button>
            }
        >
            <div className="max-w-4xl mx-auto px-6 py-8 w-full">
                {capabilityTypesLoader.loading ? (
                    <div className="flex items-center justify-center h-64">
                        <div className="text-center">
                            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500 mx-auto mb-4"></div>
                            <p className="text-gray-500">Loading capability types...</p>
                        </div>
                    </div>
                ) : (
                    <CapabilityCreateForm
                        capabilityTypes={capabilityTypesLoader.data || []}
                        defaultSpaceId={spaceId ? parseInt(spaceId) : 0}
                        onSave={handleSave}
                        onCancel={handleBack}
                    />
                )}
            </div>
        </WithAdminBodyLayout>
    );
}

const CapabilityCreateForm = ({
    capabilityTypes,
    defaultSpaceId,
    onSave,
    onCancel
}: {
    capabilityTypes: CapabilityDefinition[];
    defaultSpaceId: number;
    onSave: (data: any) => void;
    onCancel: () => void;
}) => {
    const [selectedType, setSelectedType] = useState('');
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


