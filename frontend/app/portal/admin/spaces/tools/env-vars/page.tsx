"use client";
import React, { useEffect, useState, useCallback } from 'react';
import { Edit, Trash2, Key, Database, ChevronLeft, ChevronRight, Eye, EyeOff, Copy } from 'lucide-react';
import { useSearchParams, useRouter } from 'next/navigation';
import WithAdminBodyLayout from '@/contain/Layouts/WithAdminBodyLayout';
import BigSearchBar from '@/contain/compo/BigSearchBar';
import { AddButton } from '@/contain/AddButton';
import { getPackageEnvs, updatePackageEnvs } from '@/lib/api';

/** Env vars are a flat JSON object: key -> value (single-level only). */
type EnvVarsMap = Record<string, string>;

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
    const packageId = spaceId; // install_id is the package (installed space) id
    const searchParams = useSearchParams();
    const router = useRouter();
    const [searchTerm, setSearchTerm] = useState('');
    const [editingKey, setEditingKey] = useState<string | null>(null);
    const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
    const [showSecrets, setShowSecrets] = useState<{ [key: string]: boolean }>({});
    const [envs, setEnvs] = useState<EnvVarsMap>({});
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);

    const fetchEnvs = useCallback(async () => {
        setLoading(true);
        try {
            const res = await getPackageEnvs(packageId);
            setEnvs(res.data ?? {});
        } catch (e) {
            console.error('Failed to load env vars:', e);
            setEnvs({});
        } finally {
            setLoading(false);
        }
    }, [packageId]);

    useEffect(() => {
        fetchEnvs();
    }, [fetchEnvs]);

    const persistEnvs = useCallback(async (next: EnvVarsMap) => {
        setSaving(true);
        try {
            await updatePackageEnvs(packageId, next);
            setEnvs(next);
        } catch (e) {
            console.error('Failed to save env vars:', e);
        } finally {
            setSaving(false);
        }
    }, [packageId]);

    const offset = parseInt(searchParams.get('offset') || '0', 10);
    const currentOffset = isNaN(offset) || offset < 0 ? 0 : offset;

    const entries = Object.entries(envs);
    const filteredData = searchTerm === ''
        ? entries
        : entries.filter(([k]) => k.toLowerCase().includes(searchTerm.toLowerCase()));

    const paginatedData = filteredData.slice(currentOffset, currentOffset + PAGE_LIMIT);
    const hasNext = filteredData.length > currentOffset + PAGE_LIMIT;
    const hasPrevious = currentOffset > 0;

    const handleNext = () => {
        const params = new URLSearchParams(searchParams.toString());
        params.set('offset', (currentOffset + PAGE_LIMIT).toString());
        router.push(`?${params.toString()}`);
    };

    const handlePrevious = () => {
        const newOffset = Math.max(0, currentOffset - PAGE_LIMIT);
        const params = new URLSearchParams(searchParams.toString());
        if (newOffset === 0) params.delete('offset');
        else params.set('offset', newOffset.toString());
        router.push(`?${params.toString()}`);
    };

    useEffect(() => {
        if (currentOffset > 0 && searchTerm) {
            const params = new URLSearchParams(searchParams.toString());
            params.delete('offset');
            router.push(`?${params.toString()}`);
        }
    }, [searchTerm]);

    const handleCreate = async (data: { key: string; value: string }) => {
        const key = data.key.trim();
        if (!key) return;
        const next = { ...envs, [key]: data.value };
        await persistEnvs(next);
        setIsCreateModalOpen(false);
    };

    const handleUpdate = async (oldKey: string, data: { key: string; value: string }) => {
        const newKey = data.key.trim();
        if (!newKey) return;
        const next = { ...envs };
        delete next[oldKey];
        next[newKey] = data.value;
        await persistEnvs(next);
        setEditingKey(null);
    };

    const handleDelete = async (key: string) => {
        const next = { ...envs };
        delete next[key];
        await persistEnvs(next);
    };

    const toggleSecretVisibility = (key: string) => {
        setShowSecrets(prev => ({ ...prev, [key]: !prev[key] }));
    };

    const copyToClipboard = (text: string) => {
        navigator.clipboard.writeText(text);
    };

    const formatValue = (key: string, value: string) => {
        if (!showSecrets[key]) return '•'.repeat(Math.min(value.length, 20));
        return value;
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
                                        <td colSpan={3} className="px-6 py-4 text-center text-gray-500">
                                            Loading...
                                        </td>
                                    </tr>
                                ) : paginatedData.length === 0 ? (
                                    <tr>
                                        <td colSpan={3} className="px-6 py-4 text-center text-gray-500">
                                            No environment variables found
                                        </td>
                                    </tr>
                                ) : (
                                    paginatedData.map(([key, value]) => (
                                        <tr key={key} className="hover:bg-gray-50">
                                            <td className="px-6 py-4 whitespace-nowrap">
                                                <div className="text-sm font-medium text-gray-900">{key}</div>
                                            </td>
                                            <td className="px-6 py-4">
                                                <div className="flex items-center gap-2">
                                                    <div className="font-mono text-sm text-gray-900 max-w-xs truncate">
                                                        {formatValue(key, value)}
                                                    </div>
                                                    <div className="flex items-center gap-1">
                                                        <button
                                                            type="button"
                                                            onClick={() => toggleSecretVisibility(key)}
                                                            className="text-gray-400 hover:text-gray-600"
                                                        >
                                                            {showSecrets[key] ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                                                        </button>
                                                        <button
                                                            type="button"
                                                            onClick={() => copyToClipboard(value)}
                                                            className="text-gray-400 hover:text-gray-600"
                                                        >
                                                            <Copy className="w-4 h-4" />
                                                        </button>
                                                    </div>
                                                </div>
                                            </td>
                                            <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                                                <div className="flex items-center gap-2">
                                                    <button
                                                        type="button"
                                                        onClick={() => setEditingKey(key)}
                                                        className="text-blue-600 hover:text-blue-900"
                                                    >
                                                        <Edit className="w-4 h-4" />
                                                    </button>
                                                    <button
                                                        type="button"
                                                        onClick={() => {
                                                            if (confirm('Are you sure you want to delete this environment variable?')) {
                                                                handleDelete(key);
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
                                    disabled={!hasPrevious || loading || saving}
                                    className={`flex items-center gap-1 px-4 py-2 text-sm font-medium rounded-lg transition-colors ${hasPrevious && !loading && !saving
                                        ? 'bg-white text-gray-700 border border-gray-300 hover:bg-gray-50'
                                        : 'bg-gray-100 text-gray-400 cursor-not-allowed'
                                        }`}
                                >
                                    <ChevronLeft className="w-4 h-4" />
                                    Previous
                                </button>
                                <button
                                    onClick={handleNext}
                                    disabled={!hasNext || loading || saving}
                                    className={`flex items-center gap-1 px-4 py-2 text-sm font-medium rounded-lg transition-colors ${hasNext && !loading && !saving
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
            {editingKey != null && envs[editingKey] !== undefined && (
                <EditEnvVarModal
                    keyName={editingKey}
                    value={envs[editingKey]}
                    onClose={() => setEditingKey(null)}
                    onSubmit={(data) => handleUpdate(editingKey, data)}
                />
            )}
        </WithAdminBodyLayout>
    );
};

const CreateEnvVarModal = ({ onClose, onSubmit }: { onClose: () => void; onSubmit: (data: { key: string; value: string }) => void }) => {
    const [formData, setFormData] = useState({ key: '', value: '' });

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
                            onChange={(e) => setFormData((p) => ({ ...p, key: e.target.value }))}
                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                            required
                        />
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Value</label>
                        <textarea
                            value={formData.value}
                            onChange={(e) => setFormData((p) => ({ ...p, value: e.target.value }))}
                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 font-mono text-sm"
                            rows={3}
                            required
                        />
                    </div>
                    <div className="flex gap-3 justify-end">
                        <button type="button" onClick={onClose} className="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-lg">
                            Cancel
                        </button>
                        <button type="submit" className="px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-lg">
                            Create
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
};

const EditEnvVarModal = ({
    keyName,
    value,
    onClose,
    onSubmit,
}: { keyName: string; value: string; onClose: () => void; onSubmit: (data: { key: string; value: string }) => void }) => {
    const [formData, setFormData] = useState({ key: keyName, value });

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
                            onChange={(e) => setFormData((p) => ({ ...p, key: e.target.value }))}
                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                            required
                        />
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Value</label>
                        <textarea
                            value={formData.value}
                            onChange={(e) => setFormData((p) => ({ ...p, value: e.target.value }))}
                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 font-mono text-sm"
                            rows={3}
                            required
                        />
                    </div>
                    <div className="flex gap-3 justify-end">
                        <button type="button" onClick={onClose} className="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-lg">
                            Cancel
                        </button>
                        <button type="submit" className="px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-lg">
                            Update
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
};