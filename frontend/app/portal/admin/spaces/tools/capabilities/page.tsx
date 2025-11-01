"use client";
import React, { useEffect, useRef, useState } from 'react';
import { Search, Filter, Plus, Edit, Trash2, Package, Layers, Settings, Eye } from 'lucide-react';
import { useSearchParams } from 'next/navigation';
import WithAdminBodyLayout from '@/contain/Layouts/WithAdminBodyLayout';
import BigSearchBar from '@/contain/compo/BigSearchBar';
import { AddButton } from '@/contain/AddButton';
import { useGApp } from '@/hooks';
import { 
    listSpaceCapabilities, 
    SpaceCapability, 
    createSpaceCapability, 
    updateSpaceCapability, 
    deleteSpaceCapability,
    listCapabilityTypes,
    CapabilityDefinition,
    CapabilityOptionField
} from '@/lib';
import useSimpleDataLoader from '@/hooks/useSimpleDataLoader';

export default function Page() {
    const searchParams = useSearchParams();
    const installId = searchParams.get('install_id');
    const spaceId = searchParams.get('space_id');
    
    if (!installId) {
        return <div>Install ID not provided</div>;
    }

    return <CapabilitiesListingPage 
        installId={parseInt(installId)} 
        spaceId={spaceId ? parseInt(spaceId) : undefined}
    />;
}

const CapabilitiesListingPage = ({ installId, spaceId }: { installId: number; spaceId?: number }) => {
    const [searchTerm, setSearchTerm] = useState('');
    const [selectedType, setSelectedType] = useState('');
    const [selectedScope, setSelectedScope] = useState<'all' | 'package' | 'space'>('all');
    const [editingId, setEditingId] = useState<number | null>(null);
    const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
    const gapp = useGApp();

    const loader = useSimpleDataLoader<SpaceCapability[]>({
        loader: () => {
            const params: { space_id?: number; capability_type?: string } = {};
            if (selectedScope === 'package') {
                params.space_id = 0;
            } else if (selectedScope === 'space' && spaceId) {
                params.space_id = spaceId;
            } else if (selectedScope === 'space' && spaceId === undefined) {
                // If space_id not provided but trying to filter by space, don't filter
                delete params.space_id;
            }
            if (selectedType) {
                params.capability_type = selectedType;
            }
            return listSpaceCapabilities(installId, params.space_id, params.capability_type);
        },
        ready: true,
        dependencies: [selectedScope, selectedType, installId, spaceId],
    });

    const capabilityTypesLoader = useSimpleDataLoader<CapabilityDefinition[]>({
        loader: () => listCapabilityTypes(),
        ready: true,
    });

    // Filter data based on search term
    const filteredData = loader.data?.filter(cap => {
        const matchesSearch = searchTerm === '' || 
            cap.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
            cap.capability_type.toLowerCase().includes(searchTerm.toLowerCase());
        
        return matchesSearch;
    }) || [];

    // Get unique capability types for filter dropdown
    const uniqueTypes = Array.from(new Set(loader.data?.map(cap => cap.capability_type) || []));

    const handleCreate = async (data: {
        name: string;
        capability_type: string;
        space_id?: number;
        options?: any;
        extrameta?: any;
    }) => {
        try {
            await createSpaceCapability(installId, data);
            loader.reload();
            setIsCreateModalOpen(false);
        } catch (error) {
            console.error('Failed to create capability:', error);
            alert('Failed to create capability: ' + (error as any)?.response?.data?.error || (error as any)?.message);
        }
    };

    const handleUpdate = async (id: number, data: {
        name?: string;
        capability_type?: string;
        space_id?: number;
        options?: any;
        extrameta?: any;
    }) => {
        try {
            await updateSpaceCapability(installId, id, data);
            loader.reload();
            setEditingId(null);
        } catch (error) {
            console.error('Failed to update capability:', error);
            alert('Failed to update capability: ' + (error as any)?.response?.data?.error || (error as any)?.message);
        }
    };

    const handleDelete = async (id: number) => {
        if (!confirm('Are you sure you want to delete this capability?')) {
            return;
        }
        try {
            await deleteSpaceCapability(installId, id);
            loader.reload();
        } catch (error) {
            console.error('Failed to delete capability:', error);
            alert('Failed to delete capability: ' + (error as any)?.response?.data?.error || (error as any)?.message);
        }
    };

    return (
        <WithAdminBodyLayout
            Icon={Settings}
            name="Space Capabilities"
            description="Manage capabilities for this package or space"
            rightContent={
                <AddButton
                    name="+ Add Capability"
                    onClick={() => setIsCreateModalOpen(true)}
                />
            }
        >
            <BigSearchBar
                searchText={searchTerm}
                setSearchText={setSearchTerm}
            />

            <div className="max-w-7xl mx-auto px-6 py-8 w-full">
                {/* Filters */}
                <div className="mb-6 flex gap-4 items-center flex-wrap">
                    <div className="flex items-center gap-2">
                        <Filter className="w-4 h-4" />
                        <span className="text-sm font-medium">Filter:</span>
                    </div>
                    <select
                        value={selectedType}
                        onChange={(e) => setSelectedType(e.target.value)}
                        className="px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                    >
                        <option value="">All Types</option>
                        {uniqueTypes.map(type => (
                            <option key={type} value={type}>{type}</option>
                        ))}
                    </select>
                    <select
                        value={selectedScope}
                        onChange={(e) => setSelectedScope(e.target.value as 'all' | 'package' | 'space')}
                        className="px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                    >
                        <option value="all">All Scopes</option>
                        <option value="package">Package Level (root)</option>
                        {spaceId !== undefined && <option value="space">This Space</option>}
                    </select>
                </div>

                {/* Table */}
                <div className="bg-white rounded-lg shadow overflow-hidden">
                    <div className="overflow-x-auto">
                        <table className="min-w-full divide-y divide-gray-200">
                            <thead className="bg-gray-50">
                                <tr>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                        Name
                                    </th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                        Type
                                    </th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                        Scope
                                    </th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                        Options
                                    </th>
                                    <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                                        Actions
                                    </th>
                                </tr>
                            </thead>
                            <tbody className="bg-white divide-y divide-gray-200">
                                {loader.loading ? (
                                    <tr>
                                        <td colSpan={5} className="px-6 py-4 text-center text-gray-500">
                                            Loading...
                                        </td>
                                    </tr>
                                ) : filteredData.length === 0 ? (
                                    <tr>
                                        <td colSpan={5} className="px-6 py-4 text-center text-gray-500">
                                            No capabilities found
                                        </td>
                                    </tr>
                                ) : (
                                    filteredData.map((cap) => (
                                        <CapabilityRow
                                            key={cap.id}
                                            capability={cap}
                                            onEdit={() => setEditingId(cap.id)}
                                            onDelete={() => handleDelete(cap.id)}
                                            onUpdate={(data) => handleUpdate(cap.id, data)}
                                            onCancelEdit={() => setEditingId(null)}
                                            isEditing={editingId === cap.id}
                                            capabilityTypes={capabilityTypesLoader.data || []}
                                        />
                                    ))
                                )}
                            </tbody>
                        </table>
                    </div>
                </div>
            </div>

            {/* Create Modal */}
            {isCreateModalOpen && (
                <CapabilityFormModal
                    capabilityTypes={capabilityTypesLoader.data || []}
                    defaultSpaceId={spaceId || 0}
                    onSave={handleCreate}
                    onCancel={() => setIsCreateModalOpen(false)}
                />
            )}
        </WithAdminBodyLayout>
    );
};

const CapabilityRow = ({ 
    capability, 
    onEdit, 
    onDelete,
    onUpdate,
    onCancelEdit,
    isEditing,
    capabilityTypes
}: { 
    capability: SpaceCapability;
    onEdit: () => void;
    onDelete: () => void;
    onUpdate: (data: any) => void;
    onCancelEdit: () => void;
    isEditing: boolean;
    capabilityTypes: CapabilityDefinition[];
}) => {
    const [editData, setEditData] = useState({
        name: capability.name,
        capability_type: capability.capability_type,
        space_id: capability.space_id,
    });

    if (isEditing) {
        const definition = capabilityTypes.find(t => t.name === capability.capability_type);
        return (
            <tr>
                <td colSpan={5} className="px-6 py-4">
                    <CapabilityEditForm
                        capability={capability}
                        definition={definition}
                        capabilityTypes={capabilityTypes}
                        onSave={(data) => {
                            onUpdate(data);
                        }}
                        onCancel={onCancelEdit}
                    />
                </td>
            </tr>
        );
    }

    let optionsDisplay = '';
    try {
        const options = JSON.parse(capability.options || '{}');
        optionsDisplay = Object.keys(options).length > 0 
            ? JSON.stringify(options, null, 2).substring(0, 100) + (JSON.stringify(options).length > 100 ? '...' : '')
            : '{}';
    } catch {
        optionsDisplay = capability.options || '{}';
    }

    return (
        <tr className="hover:bg-gray-50">
            <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                {capability.name}
            </td>
            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                {capability.capability_type}
            </td>
            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                {capability.space_id === 0 ? (
                    <span className="inline-flex items-center gap-1">
                        <Package className="w-4 h-4" />
                        Package Level
                    </span>
                ) : (
                    <span className="inline-flex items-center gap-1">
                        <Layers className="w-4 h-4" />
                        Space #{capability.space_id}
                    </span>
                )}
            </td>
            <td className="px-6 py-4 text-sm text-gray-500">
                <div className="max-w-xs truncate" title={optionsDisplay}>
                    {optionsDisplay}
                </div>
            </td>
            <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                <div className="flex justify-end gap-2">
                    <button
                        onClick={onEdit}
                        className="text-blue-600 hover:text-blue-900"
                    >
                        <Edit className="w-4 h-4" />
                    </button>
                    <button
                        onClick={onDelete}
                        className="text-red-600 hover:text-red-900"
                    >
                        <Trash2 className="w-4 h-4" />
                    </button>
                </div>
            </td>
        </tr>
    );
};

const CapabilityFormModal = ({
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

    const definition = capabilityTypes.find(t => t.name === selectedType);

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        if (!name || !selectedType) {
            alert('Name and type are required');
            return;
        }

        const options: Record<string, any> = {};
        if (definition) {
            definition.option_fields.forEach(field => {
                if (formData[field.key] !== undefined) {
                    options[field.key] = formData[field.key];
                } else if (field.default) {
                    options[field.key] = field.default;
                }
            });
        }

        onSave({
            name,
            capability_type: selectedType,
            space_id: spaceId,
            options: Object.keys(options).length > 0 ? options : undefined,
        });
    };

    return (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm transition-opacity flex items-center justify-center z-50">
            <div className="bg-white rounded-lg p-6 max-w-2xl w-full max-h-[90vh] overflow-y-auto">
                <h2 className="text-xl font-bold mb-4">Create Capability</h2>
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

                    {definition && definition.option_fields.length > 0 && (
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-2">
                                Configuration Options
                            </label>
                            <div className="space-y-3 border border-gray-200 rounded-lg p-4">
                                {definition.option_fields.map(field => (
                                    <CapabilityOptionFieldInput
                                        key={field.key}
                                        field={field}
                                        value={formData[field.key] !== undefined ? formData[field.key] : field.default}
                                        onChange={(value) => setFormData({ ...formData, [field.key]: value })}
                                    />
                                ))}
                            </div>
                        </div>
                    )}

                    <div className="flex justify-end gap-2 pt-4">
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
                            Create
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
};

const CapabilityEditForm = ({
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
        
        const options: Record<string, any> = {};
        if (currentDefinition) {
            currentDefinition.option_fields.forEach(field => {
                if (formData[field.key] !== undefined) {
                    options[field.key] = formData[field.key];
                }
            });
        }

        onSave({
            name,
            capability_type: capabilityType,
            space_id: spaceId,
            options: Object.keys(options).length > 0 ? options : undefined,
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

            {currentDefinition && currentDefinition.option_fields.length > 0 && (
                <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                        Configuration Options
                    </label>
                    <div className="space-y-3 border border-gray-200 rounded-lg p-4 bg-white">
                        {currentDefinition.option_fields.map(field => (
                            <CapabilityOptionFieldInput
                                key={field.key}
                                field={field}
                                value={formData[field.key] !== undefined ? formData[field.key] : field.default}
                                onChange={(value) => setFormData({ ...formData, [field.key]: value })}
                            />
                        ))}
                    </div>
                </div>
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

const CapabilityOptionFieldInput = ({
    field,
    value,
    onChange
}: {
    field: CapabilityOptionField;
    value: any;
    onChange: (value: any) => void;
}) => {
    const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
        let newValue: any = e.target.value;
        
        if (field.type === 'number') {
            newValue = parseFloat(newValue) || 0;
        } else if (field.type === 'boolean') {
            newValue = (e.target as HTMLInputElement).checked;
        }

        onChange(newValue);
    };

    return (
        <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
                {field.name}
                {field.required && <span className="text-red-500">*</span>}
            </label>
            {field.description && (
                <p className="text-xs text-gray-500 mb-1">{field.description}</p>
            )}
            {field.type === 'textarea' ? (
                <textarea
                    value={value || ''}
                    onChange={handleChange}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                    required={field.required}
                    rows={3}
                />
            ) : field.type === 'select' ? (
                <select
                    value={value || field.default || ''}
                    onChange={handleChange}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                    required={field.required}
                >
                    {field.options.map(opt => (
                        <option key={opt} value={opt}>{opt}</option>
                    ))}
                </select>
            ) : field.type === 'boolean' ? (
                <label className="flex items-center gap-2">
                    <input
                        type="checkbox"
                        checked={value || false}
                        onChange={handleChange}
                        className="w-4 h-4"
                    />
                    <span className="text-sm text-gray-700">Enabled</span>
                </label>
            ) : (
                <input
                    type={field.type === 'api_key' ? 'password' : field.type === 'number' ? 'number' : 'text'}
                    value={value || ''}
                    onChange={handleChange}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                    required={field.required}
                    placeholder={field.default}
                />
            )}
        </div>
    );
};
