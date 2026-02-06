"use client";
import React, { useEffect, useState } from 'react';
import { Filter, Edit, Trash2, Key, Database, Plus, ChevronLeft, ChevronRight, Eye, EyeOff, Copy, RotateCcw } from 'lucide-react';
import { useSearchParams, useRouter } from 'next/navigation';
import WithAdminBodyLayout from '@/contain/Layouts/WithAdminBodyLayout';
import BigSearchBar from '@/contain/compo/BigSearchBar';
import { AddButton } from '@/contain/AddButton';

interface EnvVar {
    key: string;
    value: string;
}

const PAGE_LIMIT = 50;

export default function Page() {
    const searchParams = useSearchParams();
    const installId = searchParams.get('install_id');

    if (!installId) {
        return <div>Install ID not provided</div>;
    }

    return <EnvVarsListingPage spaceId={parseInt(installId)} />;
}

const EnvVarsListingPage = ({ spaceId }: { spaceId: number }) => {
    const searchParams = useSearchParams();
    const router = useRouter();
    const [searchTerm, setSearchTerm] = useState('');
    const [selectedEnvironment, setSelectedEnvironment] = useState('');
    const [editingId, setEditingId] = useState<string | null>(null);
    const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
    const [showSecrets, setShowSecrets] = useState<{ [key: string]: boolean }>({});

    // Get offset from URL params, default to 0
    const offset = parseInt(searchParams.get('offset') || '0', 10);
    const currentOffset = isNaN(offset) || offset < 0 ? 0 : offset;

    // Mock data for now - will be replaced with actual API call
    const [mockData, setMockData] = useState<EnvVar[]>([
        {

            key: 'DATABASE_URL',
            value: 'postgresql://user:password@localhost:5432/myapp',
        },
        {

            key: 'API_BASE_URL',
            value: 'https://api.example.com',
        },
        {

            key: 'REDIS_URL',
            value: 'redis://localhost:6379',
        }
    ]);

    const [loading, setLoading] = useState(false);

    // Filter data based on search term and environment
    const filteredData = mockData?.filter(envVar => {
        const matchesSearch = searchTerm === '' ||
            envVar.key.toLowerCase().includes(searchTerm.toLowerCase());

        return matchesSearch;
    }) || [];



    // Pagination for mock data
    const paginatedData = filteredData.slice(currentOffset, currentOffset + PAGE_LIMIT);
    const hasNext = filteredData.length > currentOffset + PAGE_LIMIT;
    const hasPrevious = currentOffset > 0;

    const handleNext = () => {
        const newOffset = currentOffset + PAGE_LIMIT;
        const params = new URLSearchParams(searchParams.toString());
        params.set('offset', newOffset.toString());
        router.push(`?${params.toString()}`);
    };

    const handlePrevious = () => {
        const newOffset = Math.max(0, currentOffset - PAGE_LIMIT);
        const params = new URLSearchParams(searchParams.toString());
        if (newOffset === 0) {
            params.delete('offset');
        } else {
            params.set('offset', newOffset.toString());
        }
        router.push(`?${params.toString()}`);
    };

    // Reset offset when filters change
    useEffect(() => {
        if (currentOffset > 0 && (searchTerm || selectedEnvironment)) {
            const params = new URLSearchParams(searchParams.toString());
            params.delete('offset');
            router.push(`?${params.toString()}`);
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [searchTerm, selectedEnvironment]);

    const handleCreate = async (data: {
        key: string;
        value: string;
        description?: string;
        environment: 'development' | 'staging' | 'production';
        isSecret: boolean;
    }) => {
        try {
            // Mock creation - replace with actual API call
            const newEnvVar: EnvVar = {
                ...data,
            };
            setMockData([...mockData, newEnvVar]);
            setIsCreateModalOpen(false);
        } catch (error) {
            console.error('Failed to create environment variable:', error);
        }
    };

    const handleUpdate = async (id: string, data: {
        key?: string;
        value?: string;
        description?: string;
        environment?: 'development' | 'staging' | 'production';
        isSecret?: boolean;
    }) => {

    };

    const handleDelete = async (id: string) => {
        try {
            // Mock deletion - replace with actual API call
            setMockData(mockData.filter(envVar => envVar.key !== id));
        } catch (error) {
            console.error('Failed to delete environment variable:', error);
        }
    };

    const toggleSecretVisibility = (id: string) => {
        setShowSecrets(prev => ({ ...prev, [id]: !prev[id] }));
    };

    const copyToClipboard = (text: string) => {
        navigator.clipboard.writeText(text);
    };



    const formatValue = (envVar: EnvVar) => {
        if (!showSecrets[envVar.key]) {
            return '•'.repeat(Math.min(envVar.value.length, 20));
        }
        return envVar.value;
    };

    return (
        <WithAdminBodyLayout
            Icon={Key}
            name="Environment Variables"
            description="Manage environment variables for this space"
            variant="none"
        >
            <BigSearchBar
                searchText={searchTerm}
                setSearchText={setSearchTerm}
                rightContent={
                    <AddButton
                        name="+ Add Variable"
                        onClick={() => setIsCreateModalOpen(true)}
                    />
                }
            />

            <div className="max-w-7xl mx-auto px-6 py-8 w-full">

                {/* Table */}
                <div className="bg-white rounded-lg shadow overflow-hidden border border-gray-200">
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
                                            Value
                                        </div>
                                    </th>

                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                        Actions
                                    </th>
                                </tr>
                            </thead>
                            <tbody className="bg-white divide-y divide-gray-200">
                                {loading ? (
                                    <tr>
                                        <td colSpan={5} className="px-6 py-4 text-center text-gray-500">
                                            Loading...
                                        </td>
                                    </tr>
                                ) : paginatedData.length === 0 ? (
                                    <tr>
                                        <td colSpan={5} className="px-6 py-4 text-center text-gray-500">
                                            No environment variables found
                                        </td>
                                    </tr>
                                ) : (
                                    paginatedData.map((envVar) => (
                                        <tr
                                            key={envVar.key}
                                            className="hover:bg-gray-50 cursor-pointer"
                                            onClick={(e) => {

                                            }}
                                        >
                                            <td className="px-6 py-4 whitespace-nowrap">
                                                <div className="flex items-center gap-2">
                                                    <div className="text-sm font-medium text-gray-900">
                                                        {envVar.key}
                                                    </div>
                                                </div>
                                            </td>
                                            <td className="px-6 py-4">
                                                <div className="flex items-center gap-2">
                                                    <div className="font-mono text-sm text-gray-900 max-w-xs truncate">
                                                        {formatValue(envVar)}
                                                    </div>
                                                    <div className="flex items-center gap-1" onClick={(e) => e.stopPropagation()}>
                                                        <button
                                                            onClick={() => toggleSecretVisibility(envVar.key)}
                                                            className="text-gray-400 hover:text-gray-600"
                                                        >
                                                            {showSecrets[envVar.key] ? (
                                                                <EyeOff className="w-4 h-4" />
                                                            ) : (
                                                                <Eye className="w-4 h-4" />
                                                            )}
                                                        </button>
                                                        <button
                                                            onClick={() => copyToClipboard(envVar.value)}
                                                            className="text-gray-400 hover:text-gray-600"
                                                        >
                                                            <Copy className="w-4 h-4" />
                                                        </button>
                                                    </div>
                                                </div>
                                            </td>

                                            <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                                                <div className="flex items-center gap-2" onClick={(e) => e.stopPropagation()}>
                                                    <button
                                                        onClick={() => setEditingId(envVar.key)}
                                                        className="text-blue-600 hover:text-blue-900"
                                                    >
                                                        <Edit className="w-4 h-4" />
                                                    </button>
                                                    <button
                                                        onClick={() => {
                                                            if (confirm('Are you sure you want to delete this environment variable?')) {
                                                                handleDelete(envVar.key);
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

                    {/* Pagination Controls */}
                    {(hasPrevious || hasNext) && (
                        <div className="px-6 py-4 border-t border-gray-200 flex items-center justify-between">
                            <div className="text-sm text-gray-700">
                                Showing {currentOffset + 1} - {currentOffset + paginatedData.length} of {filteredData.length} variables
                            </div>
                            <div className="flex items-center gap-2">
                                <button
                                    onClick={handlePrevious}
                                    disabled={!hasPrevious || loading}
                                    className={`flex items-center gap-1 px-4 py-2 text-sm font-medium rounded-lg transition-colors ${hasPrevious && !loading
                                        ? 'bg-white text-gray-700 border border-gray-300 hover:bg-gray-50'
                                        : 'bg-gray-100 text-gray-400 cursor-not-allowed'
                                        }`}
                                >
                                    <ChevronLeft className="w-4 h-4" />
                                    Previous
                                </button>
                                <button
                                    onClick={handleNext}
                                    disabled={!hasNext || loading}
                                    className={`flex items-center gap-1 px-4 py-2 text-sm font-medium rounded-lg transition-colors ${hasNext && !loading
                                        ? 'bg-white text-gray-700 border border-gray-300 hover:bg-gray-50'
                                        : 'bg-gray-100 text-gray-400 cursor-not-allowed'
                                        }`}
                                >
                                    Next
                                    <ChevronRight className="w-4 h-4" />
                                </button>
                            </div>
                        </div>
                    )}
                </div>
            </div>

            {/* Create Modal */}
            {isCreateModalOpen && (
                <CreateEnvVarModal
                    onClose={() => setIsCreateModalOpen(false)}
                    onSubmit={handleCreate}
                />
            )}

            {/* Edit Modal */}
            {editingId && (
                <EditEnvVarModal
                    envVar={mockData.find(v => v.key === editingId)!}
                    onClose={() => setEditingId(null)}
                    onSubmit={(data) => handleUpdate(editingId, data)}
                />
            )}
        </WithAdminBodyLayout>
    );
};

const CreateEnvVarModal = ({ onClose, onSubmit }: { onClose: () => void; onSubmit: (data: any) => void }) => {
    const [formData, setFormData] = useState({
        key: '',
        value: '',
        description: '',
        environment: 'development' as 'development' | 'staging' | 'production',
        isSecret: false,
    });

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        onSubmit(formData);
    };

    return (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm transition-opacity flex items-center justify-center z-50">
            <div className="bg-white rounded-lg p-6 w-full max-w-md">
                <h3 className="text-lg font-semibold mb-4">Add Environment Variable</h3>
                <form onSubmit={handleSubmit} className="space-y-4">
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Key</label>
                        <input
                            type="text"
                            value={formData.key}
                            onChange={(e) => setFormData({ ...formData, key: e.target.value })}
                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 uppercase"
                            required
                        />
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Value</label>
                        <textarea
                            value={formData.value}
                            onChange={(e) => setFormData({ ...formData, value: e.target.value })}
                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 font-mono text-sm"
                            rows={3}
                            required
                        />
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
                        <input
                            type="text"
                            value={formData.description}
                            onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                        />
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Environment</label>
                        <select
                            value={formData.environment}
                            onChange={(e) => setFormData({ ...formData, environment: e.target.value as any })}
                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                        >
                            <option value="development">Development</option>
                            <option value="staging">Staging</option>
                            <option value="production">Production</option>
                        </select>
                    </div>
                    <div className="flex items-center">
                        <input
                            type="checkbox"
                            id="isSecret"
                            checked={formData.isSecret}
                            onChange={(e) => setFormData({ ...formData, isSecret: e.target.checked })}
                            className="mr-2"
                        />
                        <label htmlFor="isSecret" className="text-sm font-medium text-gray-700">
                            Mark as secret (hide value by default)
                        </label>
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

const EditEnvVarModal = ({ envVar, onClose, onSubmit }: { envVar: EnvVar; onClose: () => void; onSubmit: (data: any) => void }) => {
    const [formData, setFormData] = useState({
        key: envVar.key,
        value: envVar.value,

    });

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        onSubmit(formData);
    };

    return (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm transition-opacity flex items-center justify-center z-50">
            <div className="bg-white rounded-lg p-6 w-full max-w-md">
                <h3 className="text-lg font-semibold mb-4">Edit Environment Variable</h3>
                <form onSubmit={handleSubmit} className="space-y-4">
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Key</label>
                        <input
                            type="text"
                            value={formData.key}
                            onChange={(e) => setFormData({ ...formData, key: e.target.value })}
                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 uppercase"
                            required
                        />
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Value</label>
                        <textarea
                            value={formData.value}
                            onChange={(e) => setFormData({ ...formData, value: e.target.value })}
                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 font-mono text-sm"
                            rows={3}
                            required
                        />
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