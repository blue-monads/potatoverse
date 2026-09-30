"use client";
import React, { useState } from 'react';
import { Filter, Edit, Trash2, Users, Package, Layers, User as UserIcon, Shield } from 'lucide-react';
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
    User,
    listSpaceUserGroups,
    SpaceUserGroup,
    createSpaceUserGroup,
    updateSpaceUserGroup,
    deleteSpaceUserGroup,
    getUserGroups,
    UserGroup,
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
    const [activeTab, setActiveTab] = useState<'users' | 'groups'>('users');
    const [searchTerm, setSearchTerm] = useState('');
    const [selectedScope, setSelectedScope] = useState<'all' | 'package' | 'space'>('all');
    const [editingUserId, setEditingUserId] = useState<number | null>(null);
    const [isCreateUserModalOpen, setIsCreateUserModalOpen] = useState(false);
    const [editingGroupId, setEditingGroupId] = useState<number | null>(null);
    const [isCreateGroupModalOpen, setIsCreateGroupModalOpen] = useState(false);
    const gapp = useGApp();

    // Space Users loader
    const usersListLoader = useSimpleDataLoader<SpaceUser[]>({
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

    // Space User Groups loader
    const groupsListLoader = useSimpleDataLoader<SpaceUserGroup[]>({
        loader: () => {
            const params: { space_id?: number } = {};
            if (selectedScope === 'package') {
                params.space_id = 0;
            } else if (selectedScope === 'space' && spaceId) {
                params.space_id = spaceId;
            }
            return listSpaceUserGroups(installId, params.space_id);
        },
        ready: true,
        dependencies: [selectedScope, installId, spaceId],
    });

    const usersLoader = useSimpleDataLoader<User[]>({
        loader: () => getUsers(),
        ready: true,
    });

    const userGroupsLoader = useSimpleDataLoader<UserGroup[]>({
        loader: () => getUserGroups(),
        ready: true,
    });

    // Filter space users based on search term
    const filteredUsers = usersListLoader.data?.filter(su => {
        const user = usersLoader.data?.find(u => u.id === su.user_id);
        const matchesSearch = searchTerm === '' ||
            (user?.name && user.name.toLowerCase().includes(searchTerm.toLowerCase())) ||
            (user?.email && user.email.toLowerCase().includes(searchTerm.toLowerCase())) ||
            (su.scope && su.scope.toLowerCase().includes(searchTerm.toLowerCase())) ||
            String(su.user_id).includes(searchTerm) ||
            String(su.space_id).includes(searchTerm);

        return matchesSearch;
    }) || [];

    // Filter space groups based on search term
    const filteredGroups = groupsListLoader.data?.filter(sg => {
        const group = userGroupsLoader.data?.find(g => g.id === sg.group_id);
        const matchesSearch = searchTerm === '' ||
            (group?.name && group.name.toLowerCase().includes(searchTerm.toLowerCase())) ||
            (group?.info && group.info.toLowerCase().includes(searchTerm.toLowerCase())) ||
            (sg.scope && sg.scope.toLowerCase().includes(searchTerm.toLowerCase())) ||
            String(sg.group_id).includes(searchTerm) ||
            String(sg.space_id).includes(searchTerm);

        return matchesSearch;
    }) || [];

    // User Handlers
    const handleCreateUser = async (data: {
        user_id: number;
        space_id?: number;
        scope?: string;
    }) => {
        try {
            await createSpaceUser(installId, data);
            usersListLoader.reload();
            setIsCreateUserModalOpen(false);
        } catch (error) {
            console.error('Failed to create space user:', error);
            alert('Failed to create space user: ' + ((error as any)?.response?.data?.error || (error as any)?.message));
        }
    };

    const handleUpdateUser = async (id: number, data: {
        scope?: string;
    }) => {
        try {
            await updateSpaceUser(installId, id, data);
            usersListLoader.reload();
            setEditingUserId(null);
        } catch (error) {
            console.error('Failed to update space user:', error);
            alert('Failed to update space user: ' + ((error as any)?.response?.data?.error || (error as any)?.message));
        }
    };

    const handleDeleteUser = async (id: number) => {
        try {
            await deleteSpaceUser(installId, id);
            usersListLoader.reload();
        } catch (error) {
            console.error('Failed to delete space user:', error);
            alert('Failed to delete space user: ' + ((error as any)?.response?.data?.error || (error as any)?.message));
        }
    };

    // Group Handlers
    const handleCreateGroup = async (data: {
        group_id: number;
        space_id?: number;
        scope?: string;
    }) => {
        try {
            await createSpaceUserGroup(installId, data);
            groupsListLoader.reload();
            setIsCreateGroupModalOpen(false);
        } catch (error) {
            console.error('Failed to create space user group:', error);
            alert('Failed to create space user group: ' + ((error as any)?.response?.data?.error || (error as any)?.message));
        }
    };

    const handleUpdateGroup = async (id: number, data: {
        scope?: string;
    }) => {
        try {
            await updateSpaceUserGroup(installId, id, data);
            groupsListLoader.reload();
            setEditingGroupId(null);
        } catch (error) {
            console.error('Failed to update space user group:', error);
            alert('Failed to update space user group: ' + ((error as any)?.response?.data?.error || (error as any)?.message));
        }
    };

    const handleDeleteGroup = async (id: number) => {
        try {
            await deleteSpaceUserGroup(installId, id);
            groupsListLoader.reload();
        } catch (error) {
            console.error('Failed to delete space user group:', error);
            alert('Failed to delete space user group: ' + ((error as any)?.response?.data?.error || (error as any)?.message));
        }
    };

    const getUserInfo = (userId: number) => {
        return usersLoader.data?.find(u => u.id === userId);
    };

    const getGroupInfo = (groupId: number) => {
        return userGroupsLoader.data?.find(g => g.id === groupId);
    };

    return (
        <WithAdminBodyLayout
            Icon={UserIcon}
            name="Space Users & Groups"
            description="Manage users and user groups assigned to this package or space"
            variant="none"
        >
            <BigSearchBar
                searchText={searchTerm}
                setSearchText={setSearchTerm}
                rightContent={
                    activeTab === 'users' ? (
                        <AddButton
                            name="+ Add User"
                            onClick={() => setIsCreateUserModalOpen(true)}
                        />
                    ) : (
                        <AddButton
                            name="+ Add User Group"
                            onClick={() => setIsCreateGroupModalOpen(true)}
                        />
                    )
                }
            />

            <div className="max-w-7xl mx-auto px-6 py-8 w-full">
                {/* Tabs */}
                <div className="flex border-b border-gray-200 mb-6">
                    <button
                        onClick={() => {
                            setActiveTab('users');
                            setSearchTerm('');
                        }}
                        className={`flex items-center gap-2 px-6 py-3 font-medium text-sm border-b-2 transition-colors cursor-pointer ${
                            activeTab === 'users'
                                ? 'border-blue-600 text-blue-600'
                                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                        }`}
                    >
                        <UserIcon className="w-4 h-4" />
                        Users ({filteredUsers.length})
                    </button>
                    <button
                        onClick={() => {
                            setActiveTab('groups');
                            setSearchTerm('');
                        }}
                        className={`flex items-center gap-2 px-6 py-3 font-medium text-sm border-b-2 transition-colors cursor-pointer ${
                            activeTab === 'groups'
                                ? 'border-blue-600 text-blue-600'
                                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                        }`}
                    >
                        <Users className="w-4 h-4" />
                        User Groups ({filteredGroups.length})
                    </button>
                </div>

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

                {/* Users Table */}
                {activeTab === 'users' && (
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
                                    {usersListLoader.loading ? (
                                        <tr>
                                            <td colSpan={5} className="px-6 py-4 text-center text-gray-500">
                                                Loading...
                                            </td>
                                        </tr>
                                    ) : filteredUsers.length === 0 ? (
                                        <tr>
                                            <td colSpan={5} className="px-6 py-4 text-center text-gray-500">
                                                No space users found
                                            </td>
                                        </tr>
                                    ) : (
                                        filteredUsers.map((su) => {
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
                                                        <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${su.space_id === 0
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
                                                                onClick={() => setEditingUserId(su.id)}
                                                                className="text-blue-600 hover:text-blue-900"
                                                            >
                                                                <Edit className="w-4 h-4" />
                                                            </button>
                                                            <button
                                                                onClick={() => {
                                                                    if (confirm('Are you sure you want to remove this user from the space?')) {
                                                                        handleDeleteUser(su.id);
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
                )}

                {/* User Groups Table */}
                {activeTab === 'groups' && (
                    <div className="bg-white rounded-lg shadow overflow-hidden">
                        <div className="overflow-x-auto">
                            <table className="min-w-full divide-y divide-gray-200">
                                <thead className="bg-gray-50">
                                    <tr>
                                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                            User Group
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
                                    {groupsListLoader.loading ? (
                                        <tr>
                                            <td colSpan={5} className="px-6 py-4 text-center text-gray-500">
                                                Loading...
                                            </td>
                                        </tr>
                                    ) : filteredGroups.length === 0 ? (
                                        <tr>
                                            <td colSpan={5} className="px-6 py-4 text-center text-gray-500">
                                                No space user groups found
                                            </td>
                                        </tr>
                                    ) : (
                                        filteredGroups.map((sug) => {
                                            const group = getGroupInfo(sug.group_id);
                                            return (
                                                <tr key={sug.id} className="hover:bg-gray-50">
                                                    <td className="px-6 py-4 whitespace-nowrap">
                                                        <div className="flex items-center gap-3">
                                                            <div className="w-8 h-8 rounded-full bg-emerald-100 text-emerald-700 flex items-center justify-center font-bold text-xs">
                                                                <Users className="w-4 h-4" />
                                                            </div>
                                                            <div>
                                                                <div className="text-sm font-medium text-gray-900">
                                                                    {group?.name || `Group ID: ${sug.group_id}`}
                                                                </div>
                                                                <div className="text-sm text-gray-500">
                                                                    {group?.info || `ID: ${sug.group_id}`}
                                                                </div>
                                                            </div>
                                                        </div>
                                                    </td>
                                                    <td className="px-6 py-4 whitespace-nowrap">
                                                        <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${sug.space_id === 0
                                                            ? 'bg-purple-100 text-purple-800'
                                                            : 'bg-blue-100 text-blue-800'
                                                            }`}>
                                                            {sug.space_id === 0 ? 'Package (Root)' : 'Space'}
                                                        </span>
                                                    </td>
                                                    <td className="px-6 py-4 whitespace-nowrap">
                                                        <div className="text-sm text-gray-900">
                                                            {sug.space_id === 0 ? '-' : sug.space_id}
                                                        </div>
                                                    </td>
                                                    <td className="px-6 py-4 whitespace-nowrap">
                                                        <div className="text-sm text-gray-900">
                                                            {sug.scope || '-'}
                                                        </div>
                                                    </td>
                                                    <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                                                        <div className="flex items-center gap-2">
                                                            <button
                                                                onClick={() => setEditingGroupId(sug.id)}
                                                                className="text-blue-600 hover:text-blue-900"
                                                            >
                                                                <Edit className="w-4 h-4" />
                                                            </button>
                                                            <button
                                                                onClick={() => {
                                                                    if (confirm('Are you sure you want to remove this user group from the space?')) {
                                                                        handleDeleteGroup(sug.id);
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
                )}
            </div>

            {/* Create User Modal */}
            {isCreateUserModalOpen && (
                <CreateSpaceUserModal
                    users={usersLoader.data || []}
                    spaceId={spaceId}
                    onClose={() => setIsCreateUserModalOpen(false)}
                    onSubmit={handleCreateUser}
                />
            )}

            {/* Edit User Modal */}
            {editingUserId && (
                <EditSpaceUserModal
                    spaceUser={filteredUsers.find(su => su.id === editingUserId)!}
                    onClose={() => setEditingUserId(null)}
                    onSubmit={(data) => handleUpdateUser(editingUserId, data)}
                />
            )}

            {/* Create Group Modal */}
            {isCreateGroupModalOpen && (
                <CreateSpaceUserGroupModal
                    groups={userGroupsLoader.data || []}
                    spaceId={spaceId}
                    onClose={() => setIsCreateGroupModalOpen(false)}
                    onSubmit={handleCreateGroup}
                />
            )}

            {/* Edit Group Modal */}
            {editingGroupId && (
                <EditSpaceUserGroupModal
                    spaceUserGroup={filteredGroups.find(sg => sg.id === editingGroupId)!}
                    onClose={() => setEditingGroupId(null)}
                    onSubmit={(data) => handleUpdateGroup(editingGroupId, data)}
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
                            className="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-lg cursor-pointer"
                        >
                            Cancel
                        </button>
                        <button
                            type="submit"
                            className="px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-lg cursor-pointer"
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
                            className="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-lg cursor-pointer"
                        >
                            Cancel
                        </button>
                        <button
                            type="submit"
                            className="px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-lg cursor-pointer"
                        >
                            Update
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
};

const CreateSpaceUserGroupModal = ({
    groups,
    spaceId,
    onClose,
    onSubmit
}: {
    groups: UserGroup[];
    spaceId?: number;
    onClose: () => void;
    onSubmit: (data: any) => void;
}) => {
    const [formData, setFormData] = useState({
        group_id: '',
        space_id: spaceId || 0,
        scope: '',
    });

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        onSubmit({
            group_id: parseInt(formData.group_id),
            space_id: formData.space_id === 0 ? undefined : formData.space_id,
            scope: formData.scope || undefined,
        });
    };

    return (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm transition-opacity flex items-center justify-center z-50">
            <div className="bg-white rounded-lg p-6 w-full max-w-md">
                <h3 className="text-lg font-semibold mb-4">Add User Group to Space/Package</h3>
                <form onSubmit={handleSubmit} className="space-y-4">
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">User Group</label>
                        <select
                            value={formData.group_id}
                            onChange={(e) => setFormData({ ...formData, group_id: e.target.value })}
                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                            required
                        >
                            <option value="">Select a user group</option>
                            {groups.map(group => (
                                <option key={group.id ?? group.name} value={group.id}>
                                    {group.name} {group.info ? `(${group.info})` : ''}
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
                            className="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-lg cursor-pointer"
                        >
                            Cancel
                        </button>
                        <button
                            type="submit"
                            className="px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-lg cursor-pointer"
                        >
                            Add User Group
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
};

const EditSpaceUserGroupModal = ({
    spaceUserGroup,
    onClose,
    onSubmit
}: {
    spaceUserGroup: SpaceUserGroup;
    onClose: () => void;
    onSubmit: (data: any) => void;
}) => {
    const [formData, setFormData] = useState({
        scope: spaceUserGroup.scope || '',
    });

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        onSubmit(formData);
    };

    return (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm transition-opacity flex items-center justify-center z-50">
            <div className="bg-white rounded-lg p-6 w-full max-w-md">
                <h3 className="text-lg font-semibold mb-4">Edit Space User Group</h3>
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
                            className="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-lg cursor-pointer"
                        >
                            Cancel
                        </button>
                        <button
                            type="submit"
                            className="px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-lg cursor-pointer"
                        >
                            Update
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
};
