"use client";
import React, { useState } from 'react';
import { Search, Filter, Edit, Trash2, Users, Package, Layers, UserPlus } from 'lucide-react';
import { useSearchParams } from 'next/navigation';
import WithAdminBodyLayout from '@/contain/Layouts/WithAdminBodyLayout';
import BigSearchBar from '@/contain/compo/BigSearchBar';
import { AddButton } from '@/contain/AddButton';
import { useGApp } from '@/hooks';
import { 
    listSpaceUsers, 
    SpaceUser, 
    createSpaceUser, 
    updateSpaceUser, 
    deleteSpaceUser,
    getUsers,
    User
} from '@/lib';
import useSimpleDataLoader from '@/hooks/useSimpleDataLoader';

export default function Page() {
    const searchParams = useSearchParams();
    const installId = searchParams.get('install_id');
    const spaceId = searchParams.get('space_id');
    
    if (!installId) {
        return <div>Install ID not provided</div>;
    }

    return <SpaceUsersListingPage 
        installId={parseInt(installId)} 
        spaceId={spaceId ? parseInt(spaceId) : undefined}
    />;
}

const SpaceUsersListingPage = ({ installId, spaceId }: { installId: number; spaceId?: number }) => {
    const [searchTerm, setSearchTerm] = useState('');
    const [selectedScope, setSelectedScope] = useState<'all' | 'package' | 'space'>('all');
    const [editingId, setEditingId] = useState<number | null>(null);
    const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
    const gapp = useGApp();

    const loader = useSimpleDataLoader<SpaceUser[]>({
        loader: () => {
            const params: { space_id?: number } = {};
            if (selectedScope === 'package') {
                params.space_id = 0;
            } else if (selectedScope === 'space' && spaceId) {
                params.space_id = spaceId;
            }
            return listSpaceUsers(installId, params.space_id);
        },
        ready: true,
        dependencies: [selectedScope, installId, spaceId],
    });

    const usersLoader = useSimpleDataLoader<User[]>({
        loader: () => getUsers(),
        ready: true,
    });

    // Filter data based on search term
    const filteredData = loader.data?.filter(su => {
        const user = usersLoader.data?.find(u => u.id === su.user_id);
        const matchesSearch = searchTerm === '' || 
            (user?.name && user.name.toLowerCase().includes(searchTerm.toLowerCase())) ||
            (user?.email && user.email.toLowerCase().includes(searchTerm.toLowerCase())) ||
            su.scope.toLowerCase().includes(searchTerm.toLowerCase()) ||
            String(su.user_id).includes(searchTerm) ||
            String(su.space_id).includes(searchTerm);
        
        return matchesSearch;
    }) || [];

    const handleCreate = async (data: {
        user_id: number;
        space_id?: number;
        scope?: string;
    }) => {
        try {
            await createSpaceUser(installId, data);
            loader.reload();
            setIsCreateModalOpen(false);
        } catch (error) {
            console.error('Failed to create space user:', error);
            alert('Failed to create space user: ' + ((error as any)?.response?.data?.error || (error as any)?.message));
        }
    };

    const handleUpdate = async (id: number, data: {
        scope?: string;
    }) => {
        try {
            await updateSpaceUser(installId, id, data);
            loader.reload();
            setEditingId(null);
        } catch (error) {
            console.error('Failed to update space user:', error);
            alert('Failed to update space user: ' + ((error as any)?.response?.data?.error || (error as any)?.message));
        }
    };

    const handleDelete = async (id: number) => {
        try {
            await deleteSpaceUser(installId, id);
            loader.reload();
        } catch (error) {
            console.error('Failed to delete space user:', error);
            alert('Failed to delete space user: ' + ((error as any)?.response?.data?.error || (error as any)?.message));
        }
    };

    const getUserInfo = (userId: number) => {
        return usersLoader.data?.find(u => u.id === userId);
    };

    return (
        <WithAdminBodyLayout
            Icon={Users}
            name="Space Users"
            description="Manage users assigned to this package or space"
            rightContent={
                <AddButton
                    name="+ Add User"
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
                        <span className="text-sm font-medium">Filter by Scope:</span>
                    </div>
                    <select
                        value={selectedScope}
                        onChange={(e) => setSelectedScope(e.target.value as 'all' | 'package' | 'space')}
                        className="px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                    >
                        <option value="all">All</option>
                        <option value="package">Package Level (Root)</option>
                        {spaceId && <option value="space">Space Level</option>}
                    </select>
                </div>

                {/* Table */}
                <div className="bg-white rounded-lg shadow overflow-hidden">
                    <div className="overflow-x-auto">
                        <table className="min-w-full divide-y divide-gray-200">
                            <thead className="bg-gray-50">
                                <tr>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                        User
                                    </th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                        <div className="flex items-center gap-2">
                                            <Package className="w-4 h-4" />
                                            Scope
                                        </div>
                                    </th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                        <div className="flex items-center gap-2">
                                            <Layers className="w-4 h-4" />
                                            Space ID
                                        </div>
                                    </th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                        Scope/Permission
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
                                            No space users found
                                        </td>
                                    </tr>
                                ) : (
                                    filteredData.map((su) => {
                                        const user = getUserInfo(su.user_id);
                                        return (
                                            <tr key={su.id} className="hover:bg-gray-50">
                                                <td className="px-6 py-4 whitespace-nowrap">
                                                    <div className="flex items-center gap-3">
                                                        <img 
                                                            src={`/zz/profileImage/${su.user_id}/${user?.name || 'User'}`} 
                                                            alt="profile" 
                                                            className="w-8 h-8 rounded-full"
                                                        />
                                                        <div>
                                                            <div className="text-sm font-medium text-gray-900">
                                                                {user?.name || `User ${su.user_id}`}
                                                            </div>
                                                            <div className="text-sm text-gray-500">
                                                                {user?.email || `ID: ${su.user_id}`}
                                                            </div>
                                                        </div>
                                                    </div>
                                                </td>
                                                <td className="px-6 py-4 whitespace-nowrap">
                                                    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                                                        su.space_id === 0 
                                                            ? 'bg-purple-100 text-purple-800' 
                                                            : 'bg-blue-100 text-blue-800'
                                                    }`}>
                                                        {su.space_id === 0 ? 'Package (Root)' : 'Space'}
                                                    </span>
                                                </td>
                                                <td className="px-6 py-4 whitespace-nowrap">
                                                    <div className="text-sm text-gray-900">
                                                        {su.space_id === 0 ? '-' : su.space_id}
                                                    </div>
                                                </td>
                                                <td className="px-6 py-4 whitespace-nowrap">
                                                    <div className="text-sm text-gray-900">
                                                        {su.scope || '-'}
                                                    </div>
                                                </td>
                                                <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                                                    <div className="flex items-center gap-2">
                                                        <button
                                                            onClick={() => setEditingId(su.id)}
                                                            className="text-blue-600 hover:text-blue-900"
                                                        >
                                                            <Edit className="w-4 h-4" />
                                                        </button>
                                                        <button
                                                            onClick={() => {
                                                                if (confirm('Are you sure you want to remove this user from the space?')) {
                                                                    handleDelete(su.id);
                                                                }
                                                            }}
                                                            className="text-red-600 hover:text-red-900"
                                                        >
                                                            <Trash2 className="w-4 h-4" />
                                                        </button>
                                                    </div>
                                                </td>
                                            </tr>
                                        );
                                    })
                                )}
                            </tbody>
                        </table>
                    </div>
                </div>
            </div>

            {/* Create Modal */}
            {isCreateModalOpen && (
                <CreateSpaceUserModal
                    users={usersLoader.data || []}
                    spaceId={spaceId}
                    onClose={() => setIsCreateModalOpen(false)}
                    onSubmit={handleCreate}
                />
            )}

            {/* Edit Modal */}
            {editingId && (
                <EditSpaceUserModal
                    spaceUser={filteredData.find(su => su.id === editingId)!}
                    onClose={() => setEditingId(null)}
                    onSubmit={(data) => handleUpdate(editingId, data)}
                />
            )}
        </WithAdminBodyLayout>
    );
};

const CreateSpaceUserModal = ({ 
    users, 
    spaceId, 
    onClose, 
    onSubmit 
}: { 
    users: User[]; 
    spaceId?: number;
    onClose: () => void; 
    onSubmit: (data: any) => void;
}) => {
    const [formData, setFormData] = useState({
        user_id: '',
        space_id: spaceId || 0,
        scope: '',
    });

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        onSubmit({
            user_id: parseInt(formData.user_id),
            space_id: formData.space_id === 0 ? undefined : formData.space_id,
            scope: formData.scope || undefined,
        });
    };

    return (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm transition-opacity flex items-center justify-center z-50">
            <div className="bg-white rounded-lg p-6 w-full max-w-md">
                <h3 className="text-lg font-semibold mb-4">Add User to Space/Package</h3>
                <form onSubmit={handleSubmit} className="space-y-4">
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">User</label>
                        <select
                            value={formData.user_id}
                            onChange={(e) => setFormData({ ...formData, user_id: e.target.value })}
                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                            required
                        >
                            <option value="">Select a user</option>
                            {users.map(user => (
                                <option key={user.id} value={user.id}>
                                    {user.name} ({user.email})
                                </option>
                            ))}
                        </select>
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Level</label>
                        <select
                            value={formData.space_id}
                            onChange={(e) => setFormData({ ...formData, space_id: parseInt(e.target.value) })}
                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                        >
                            <option value="0">Package Level (Root)</option>
                            {spaceId && <option value={spaceId}>Space Level</option>}
                        </select>
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Scope/Permission</label>
                        <input
                            type="text"
                            value={formData.scope}
                            onChange={(e) => setFormData({ ...formData, scope: e.target.value })}
                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                            placeholder="e.g., admin, user, viewer"
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
                            Add User
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
};

const EditSpaceUserModal = ({ 
    spaceUser, 
    onClose, 
    onSubmit 
}: { 
    spaceUser: SpaceUser; 
    onClose: () => void; 
    onSubmit: (data: any) => void;
}) => {
    const [formData, setFormData] = useState({
        scope: spaceUser.scope || '',
    });

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        onSubmit(formData);
    };

    return (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm transition-opacity flex items-center justify-center z-50">
            <div className="bg-white rounded-lg p-6 w-full max-w-md">
                <h3 className="text-lg font-semibold mb-4">Edit Space User</h3>
                <form onSubmit={handleSubmit} className="space-y-4">
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Scope/Permission</label>
                        <input
                            type="text"
                            value={formData.scope}
                            onChange={(e) => setFormData({ ...formData, scope: e.target.value })}
                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                            placeholder="e.g., admin, user, viewer"
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
