"use client";
import React, { useState } from 'react';
import { Filter, Edit, Trash2, Package, Layers, Settings, Bug } from 'lucide-react';
import { useSearchParams, useRouter } from 'next/navigation';
import WithAdminBodyLayout from '@/contain/Layouts/WithAdminBodyLayout';
import BigSearchBar from '@/contain/compo/BigSearchBar';
import { AddButton } from '@/contain/AddButton';
import { useGApp } from '@/hooks';
import { 
    listSpaceCapabilities, 
    SpaceCapability, 
    updateSpaceCapability, 
    deleteSpaceCapability,
    listCapabilityTypes,
    CapabilityDefinition,
    getCapabilitiesDebug
} from '@/lib';
import useSimpleDataLoader from '@/hooks/useSimpleDataLoader';
import CapEditor from './sub/CapEditor';


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
    const router = useRouter();
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

    const handleCreateClick = () => {
        const params = new URLSearchParams();
        params.set('install_id', installId.toString());
        if (spaceId) params.set('space_id', spaceId.toString());
        router.push(`/portal/admin/spaces/tools/capabilities/create?${params.toString()}`);
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
                    onClick={handleCreateClick}
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
                                            installId={installId}
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
        </WithAdminBodyLayout>
    );
};

const CapabilityRow = ({ 
    capability, 
    installId,
    onEdit, 
    onDelete,
    onUpdate,
    onCancelEdit,
    isEditing,
    capabilityTypes
}: { 
    capability: SpaceCapability;
    installId: number;
    onEdit: () => void;
    onDelete: () => void;
    onUpdate: (data: any) => void;
    onCancelEdit: () => void;
    isEditing: boolean;
    capabilityTypes: CapabilityDefinition[];
}) => {
    const gapp = useGApp();
    const [loadingDebug, setLoadingDebug] = useState(false);

    const handleDebugClick = async () => {
        setLoadingDebug(true);
        try {
            const response = await getCapabilitiesDebug(capability.capability_type);
            const debugData = response.data;
            
            gapp.modal.openModal({
                title: `Debug: ${capability.name} (${capability.capability_type})`,
                content: (
                    <div className="space-y-4">
                        <div className="bg-gray-50 rounded-lg p-4 overflow-auto max-h-[70vh]">
                            <pre className="text-sm text-gray-800 whitespace-pre-wrap break-words">
                                {JSON.stringify(debugData, null, 2)}
                            </pre>
                        </div>
                        <div className="flex justify-end">
                            <button
                                onClick={() => gapp.modal.closeModal()}
                                className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600"
                            >
                                Close
                            </button>
                        </div>
                    </div>
                ),
                size: "lg"
            });
        } catch (error) {
            console.error('Failed to fetch debug data:', error);
            alert('Failed to fetch debug data: ' + ((error as any)?.response?.data?.error || (error as any)?.message || 'Unknown error'));
        } finally {
            setLoadingDebug(false);
        }
    };

    if (isEditing) {
        const definition = capabilityTypes.find(t => t.name === capability.capability_type);
        return (
            <tr>
                <td colSpan={5} className="px-6 py-4">
                    <CapEditor
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
                        onClick={handleDebugClick}
                        disabled={loadingDebug}
                        className="text-purple-600 hover:text-purple-900 disabled:opacity-50 disabled:cursor-not-allowed"
                        title="Show debug data"
                    >
                        <Bug className="w-4 h-4" />
                    </button>
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

