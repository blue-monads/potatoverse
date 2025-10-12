"use client";
import React, { useEffect, useRef, useState } from 'react';
import { Search, Filter, ArrowUpDown, Plus, Edit, Trash2, Eye, EyeOff, Key, Tag, Database, Grid2x2Plus } from 'lucide-react';
import { useSearchParams } from 'next/navigation';
import WithAdminBodyLayout from '@/contain/Layouts/WithAdminBodyLayout';
import BigSearchBar from '@/contain/compo/BigSearchBar';
import { AddButton } from '@/contain/AddButton';
import { GAppStateHandle, ModalHandle, useGApp } from '@/hooks';
import { listSpaceKV, SpaceKV, createSpaceKV, updateSpaceKV, deleteSpaceKV } from '@/lib';
import useSimpleDataLoader from '@/hooks/useSimpleDataLoader';

export default function Page() {
    const searchParams = useSearchParams();
    const spaceId = searchParams.get('id');
    
    if (!spaceId) {
        return <div>Space ID not provided</div>;
    }

    return <KVListingPage spaceId={parseInt(spaceId)} />;
}

const KVListingPage = ({ spaceId }: { spaceId: number }) => {
    const [searchTerm, setSearchTerm] = useState('');
    const [selectedGroup, setSelectedGroup] = useState('');
    const [showValue, setShowValue] = useState<Record<number, boolean>>({});
    const [editingId, setEditingId] = useState<number | null>(null);
    const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
    const gapp = useGApp();

    const loader = useSimpleDataLoader<SpaceKV[]>({
        loader: () => listSpaceKV(spaceId),
        ready: true,
    });

    // Filter data based on search term and group
    const filteredData = loader.data?.filter(kv => {
        const matchesSearch = searchTerm === '' || 
            kv.key.toLowerCase().includes(searchTerm.toLowerCase()) ||
            kv.value.toLowerCase().includes(searchTerm.toLowerCase()) ||
            kv.group.toLowerCase().includes(searchTerm.toLowerCase()) ||
            kv.tag1.toLowerCase().includes(searchTerm.toLowerCase()) ||
            kv.tag2.toLowerCase().includes(searchTerm.toLowerCase()) ||
            kv.tag3.toLowerCase().includes(searchTerm.toLowerCase());
        
        const matchesGroup = selectedGroup === '' || kv.group === selectedGroup;
        
        return matchesSearch && matchesGroup;
    }) || [];

    // Get unique groups for filter dropdown
    const uniqueGroups = Array.from(new Set(loader.data?.map(kv => kv.group) || []));

    const handleCreate = async (data: {
        key: string;
        group: string;
        value: string;
        tag1?: string;
        tag2?: string;
        tag3?: string;
    }) => {
        try {
            await createSpaceKV(spaceId, data);
            loader.reload();
            setIsCreateModalOpen(false);
        } catch (error) {
            console.error('Failed to create KV entry:', error);
        }
    };

    const handleUpdate = async (id: number, data: {
        key?: string;
        group?: string;
        value?: string;
        tag1?: string;
        tag2?: string;
        tag3?: string;
    }) => {
        try {
            await updateSpaceKV(spaceId, id, data);
            loader.reload();
            setEditingId(null);
        } catch (error) {
            console.error('Failed to update KV entry:', error);
        }
    };

    const handleDelete = async (id: number) => {
        try {
            await deleteSpaceKV(spaceId, id);
            loader.reload();
        } catch (error) {
            console.error('Failed to delete KV entry:', error);
        }
    };

    const toggleValueVisibility = (id: number) => {
        setShowValue(prev => ({
            ...prev,
            [id]: !prev[id]
        }));
    };

    return (
        <WithAdminBodyLayout
            Icon={Grid2x2Plus}
            name="Space KV Store"
            description="Key-Value storage for this space"
            rightContent={
                <AddButton
                    name="+ Add KV"
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
                <div className="mb-6 flex gap-4 items-center">
                    <div className="flex items-center gap-2">
                        <Filter className="w-4 h-4" />
                        <span className="text-sm font-medium">Filter by Group:</span>
                    </div>
                    <select
                        value={selectedGroup}
                        onChange={(e) => setSelectedGroup(e.target.value)}
                        className="px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                    >
                        <option value="">All Groups</option>
                        {uniqueGroups.map(group => (
                            <option key={group} value={group}>{group}</option>
                        ))}
                    </select>
                </div>

                {/* Table */}
                <div className="bg-white rounded-lg shadow overflow-hidden">
                    <div className="overflow-x-auto">
                        <table className="min-w-full divide-y divide-gray-200">
                            <thead className="bg-gray-50">
                                <tr>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                        <div className="flex items-center gap-2">
                                            <Key className="w-4 h-4" />
                                            Key
                                        </div>
                                    </th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                        <div className="flex items-center gap-2">
                                            <Database className="w-4 h-4" />
                                            Group
                                        </div>
                                    </th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                        Value
                                    </th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                        <div className="flex items-center gap-2">
                                            <Tag className="w-4 h-4" />
                                            Tags
                                        </div>
                                    </th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
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
                                            No KV entries found
                                        </td>
                                    </tr>
                                ) : (
                                    filteredData.map((kv) => (
                                        <tr key={kv.id} className="hover:bg-gray-50">
                                            <td className="px-6 py-4 whitespace-nowrap">
                                                <div className="text-sm font-medium text-gray-900">
                                                    {kv.key}
                                                </div>
                                                <div className="text-sm text-gray-500">
                                                    ID: {kv.id}
                                                </div>
                                            </td>
                                            <td className="px-6 py-4 whitespace-nowrap">
                                                <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                                                    {kv.group}
                                                </span>
                                            </td>
                                            <td className="px-6 py-4">
                                                <div className="flex items-center gap-2">
                                                    <div className="text-sm text-gray-900 max-w-xs truncate">
                                                        {showValue[kv.id] ? kv.value : '••••••••'}
                                                    </div>
                                                    <button
                                                        onClick={() => toggleValueVisibility(kv.id)}
                                                        className="text-gray-400 hover:text-gray-600"
                                                    >
                                                        {showValue[kv.id] ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                                                    </button>
                                                </div>
                                            </td>
                                            <td className="px-6 py-4 whitespace-nowrap">
                                                <div className="flex flex-wrap gap-1">
                                                    {kv.tag1 && (
                                                        <span className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-gray-100 text-gray-800">
                                                            {kv.tag1}
                                                        </span>
                                                    )}
                                                    {kv.tag2 && (
                                                        <span className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-gray-100 text-gray-800">
                                                            {kv.tag2}
                                                        </span>
                                                    )}
                                                    {kv.tag3 && (
                                                        <span className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-gray-100 text-gray-800">
                                                            {kv.tag3}
                                                        </span>
                                                    )}
                                                </div>
                                            </td>
                                            <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                                                <div className="flex items-center gap-2">
                                                    <button
                                                        onClick={() => setEditingId(kv.id)}
                                                        className="text-blue-600 hover:text-blue-900"
                                                    >
                                                        <Edit className="w-4 h-4" />
                                                    </button>
                                                    <button
                                                        onClick={() => {
                                                            if (confirm('Are you sure you want to delete this KV entry?')) {
                                                                handleDelete(kv.id);
                                                            }
                                                        }}
                                                        className="text-red-600 hover:text-red-900"
                                                    >
                                                        <Trash2 className="w-4 h-4" />
                                                    </button>
                                                </div>
                                            </td>
                                        </tr>
                                    ))
                                )}
                            </tbody>
                        </table>
                    </div>
                </div>
            </div>

            {/* Create Modal */}
            {isCreateModalOpen && (
                <CreateKVModal
                    onClose={() => setIsCreateModalOpen(false)}
                    onSubmit={handleCreate}
                />
            )}

            {/* Edit Modal */}
            {editingId && (
                <EditKVModal
                    kv={filteredData.find(kv => kv.id === editingId)!}
                    onClose={() => setEditingId(null)}
                    onSubmit={(data) => handleUpdate(editingId, data)}
                />
            )}
        </WithAdminBodyLayout>
    );
};

const CreateKVModal = ({ onClose, onSubmit }: { onClose: () => void; onSubmit: (data: any) => void }) => {
    const [formData, setFormData] = useState({
        key: '',
        group: '',
        value: '',
        tag1: '',
        tag2: '',
        tag3: '',
    });

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        onSubmit(formData);
    };

    return (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm transition-opacity flex items-center justify-center z-50">
            <div className="bg-white rounded-lg p-6 w-full max-w-md">
                <h3 className="text-lg font-semibold mb-4">Create KV Entry</h3>
                <form onSubmit={handleSubmit} className="space-y-4">
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Key</label>
                        <input
                            type="text"
                            value={formData.key}
                            onChange={(e) => setFormData({ ...formData, key: e.target.value })}
                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                            required
                        />
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Group</label>
                        <input
                            type="text"
                            value={formData.group}
                            onChange={(e) => setFormData({ ...formData, group: e.target.value })}
                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                            required
                        />
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Value</label>
                        <textarea
                            value={formData.value}
                            onChange={(e) => setFormData({ ...formData, value: e.target.value })}
                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                            rows={3}
                            required
                        />
                    </div>
                    <div className="grid grid-cols-3 gap-2">
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">Tag 1</label>
                            <input
                                type="text"
                                value={formData.tag1}
                                onChange={(e) => setFormData({ ...formData, tag1: e.target.value })}
                                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                            />
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">Tag 2</label>
                            <input
                                type="text"
                                value={formData.tag2}
                                onChange={(e) => setFormData({ ...formData, tag2: e.target.value })}
                                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                            />
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">Tag 3</label>
                            <input
                                type="text"
                                value={formData.tag3}
                                onChange={(e) => setFormData({ ...formData, tag3: e.target.value })}
                                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                            />
                        </div>
                    </div>
                    <div className="flex gap-3 justify-end">
                        <button
                            type="button"
                            onClick={onClose}
                            className="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-lg"
                        >
                            Cancel
                        </button>
                        <button
                            type="submit"
                            className="px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-lg"
                        >
                            Create
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
};

const EditKVModal = ({ kv, onClose, onSubmit }: { kv: SpaceKV; onClose: () => void; onSubmit: (data: any) => void }) => {
    const [formData, setFormData] = useState({
        key: kv.key,
        group: kv.group,
        value: kv.value,
        tag1: kv.tag1,
        tag2: kv.tag2,
        tag3: kv.tag3,
    });

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        onSubmit(formData);
    };

    return (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm transition-opacity flex items-center justify-center z-50">
            <div className="bg-white rounded-lg p-6 w-full max-w-md">
                <h3 className="text-lg font-semibold mb-4">Edit KV Entry</h3>
                <form onSubmit={handleSubmit} className="space-y-4">
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Key</label>
                        <input
                            type="text"
                            value={formData.key}
                            onChange={(e) => setFormData({ ...formData, key: e.target.value })}
                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                            required
                        />
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Group</label>
                        <input
                            type="text"
                            value={formData.group}
                            onChange={(e) => setFormData({ ...formData, group: e.target.value })}
                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                            required
                        />
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Value</label>
                        <textarea
                            value={formData.value}
                            onChange={(e) => setFormData({ ...formData, value: e.target.value })}
                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                            rows={3}
                            required
                        />
                    </div>
                    <div className="grid grid-cols-3 gap-2">
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">Tag 1</label>
                            <input
                                type="text"
                                value={formData.tag1}
                                onChange={(e) => setFormData({ ...formData, tag1: e.target.value })}
                                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                            />
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">Tag 2</label>
                            <input
                                type="text"
                                value={formData.tag2}
                                onChange={(e) => setFormData({ ...formData, tag2: e.target.value })}
                                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                            />
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">Tag 3</label>
                            <input
                                type="text"
                                value={formData.tag3}
                                onChange={(e) => setFormData({ ...formData, tag3: e.target.value })}
                                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                            />
                        </div>
                    </div>
                    <div className="flex gap-3 justify-end">
                        <button
                            type="button"
                            onClick={onClose}
                            className="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-lg"
                        >
                            Cancel
                        </button>
                        <button
                            type="submit"
                            className="px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-lg"
                        >
                            Update
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
};
