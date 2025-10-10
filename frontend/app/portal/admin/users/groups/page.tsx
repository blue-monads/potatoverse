"use client"

import { useState, useEffect } from 'react';
import { getUserGroups, createUserGroup, updateUserGroup, deleteUserGroup, UserGroup } from '@/lib/api';
import { useGApp } from '@/hooks';
import WithAdminBodyLayout from '@/contain/Layouts/WithAdminBodyLayout';
import WithTabbedUserLayout from '../WithTabbedUserLayout';
import BigSearchBar from '@/contain/compo/BigSearchBar';
import { AddButton } from '@/contain/AddButton';
import { Users, UserIcon, MoreVertical, Edit, Trash2 } from 'lucide-react';

export default function UserGroupsPage() {
    const [groups, setGroups] = useState<UserGroup[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [searchTerm, setSearchTerm] = useState('');
    const [openDropdownId, setOpenDropdownId] = useState<string | null>(null);
    const gapp = useGApp();

    const loadGroups = async () => {
        try {
            setLoading(true);
            const response = await getUserGroups();
            setGroups(response.data);
            setError(null);
        } catch (err) {
            setError('Failed to load user groups');
            console.error('Error loading groups:', err);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        if (!gapp.isInitialized) {
            return;
        }
        loadGroups();
    }, [gapp.isInitialized]);

    const handleCreateGroup = async (data: { name: string; info: string }) => {
        try {
            await createUserGroup(data);
            await loadGroups();
            gapp.modal.closeModal();
        } catch (err) {
            console.error('Error creating group:', err);
            throw err;
        }
    };

    const handleUpdateGroup = async (name: string, data: { info: string }) => {
        try {
            await updateUserGroup(name, data);
            await loadGroups();
            gapp.modal.closeModal();
        } catch (err) {
            console.error('Error updating group:', err);
            throw err;
        }
    };

    const handleDeleteGroup = async (name: string) => {
        if (!confirm(`Are you sure you want to delete the group "${name}"?`)) {
            return;
        }

        try {
            await deleteUserGroup(name);
            await loadGroups();
        } catch (err) {
            console.error('Error deleting group:', err);
            setError('Failed to delete user group');
        }
    };

    const openCreateModal = () => {
        gapp.modal.openModal({
            title: 'Create User Group',
            content: <CreateGroupForm onSubmit={handleCreateGroup} />,
            size: 'md'
        });
    };

    const openEditModal = (group: UserGroup) => {
        gapp.modal.openModal({
            title: 'Edit User Group',
            content: <EditGroupForm group={group} onSubmit={handleUpdateGroup} />,
            size: 'md'
        });
    };

    const toggleDropdown = (name: string) => {
        setOpenDropdownId(openDropdownId === name ? null : name);
    };

    const closeDropdown = () => {
        setOpenDropdownId(null);
    };

    const filteredGroups = groups.filter(group =>
        group.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        (group.info && group.info.toLowerCase().includes(searchTerm.toLowerCase()))
    );

    return (
        <WithAdminBodyLayout
            Icon={UserIcon}
            name='Users'
            description="Manage your users, roles, and permissions."
            rightContent={
                <AddButton
                    name="+ Group"
                    onClick={openCreateModal}
                />
            }
        >
            <BigSearchBar
                setSearchText={setSearchTerm}
                searchText={searchTerm}
                placeholder="Search groups..."
            />

            <WithTabbedUserLayout activeTab="groups">
                <div className="max-w-7xl mx-auto">
                    {loading ? (
                        <div className="flex justify-center items-center h-32">
                            <div className="text-gray-500">Loading groups...</div>
                        </div>
                    ) : filteredGroups.length === 0 ? (
                        <div className="text-center py-12">
                            <Users className="w-12 h-12 text-gray-400 mx-auto mb-4" />
                            <h3 className="text-lg font-medium text-gray-900 mb-2">No groups yet</h3>
                            <p className="text-gray-500 mb-4">Get started by creating your first user group.</p>
                            <button
                                className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
                                onClick={openCreateModal}
                            >
                                Create Group
                            </button>
                        </div>
                    ) : (
                        <div className="space-y-4">
                            {filteredGroups.map((group) => (
                                <div key={group.name} className="bg-white border border-gray-200 rounded-lg p-4 shadow-sm">
                                    <div className="flex items-center justify-between">
                                        <div className="flex items-center space-x-4">
                                            <div className="flex items-center space-x-2">
                                                <Users className="w-4 h-4 text-blue-500" />
                                                <div>
                                                    <div className="font-medium">{group.name}</div>
                                                    <div className="text-sm text-gray-500">
                                                        {group.info || 'No description provided'}
                                                    </div>
                                                </div>
                                            </div>
                                        </div>
                                        <div className="flex items-center space-x-3">
                                            <div className="text-sm text-gray-500">
                                                {group.created_at && new Date(group.created_at).toLocaleDateString()}
                                            </div>
                                            <div className="relative">
                                                <button 
                                                    className="p-2 hover:bg-gray-100 rounded-md"
                                                    onClick={() => toggleDropdown(group.name)}
                                                >
                                                    <MoreVertical className="w-4 h-4" />
                                                </button>
                                                {openDropdownId === group.name && (
                                                    <>
                                                        <div 
                                                            className="fixed inset-0 z-10" 
                                                            onClick={closeDropdown}
                                                        ></div>
                                                        <div className="absolute right-0 mt-2 w-48 bg-white border border-gray-200 rounded-md shadow-lg z-20">
                                                            <div className="py-1">
                                                                <button
                                                                    onClick={() => {
                                                                        openEditModal(group);
                                                                        closeDropdown();
                                                                    }}
                                                                    className="block w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                                                                >
                                                                    <Edit className="w-4 h-4 inline mr-2" />
                                                                    Edit
                                                                </button>
                                                                <button
                                                                    onClick={() => {
                                                                        handleDeleteGroup(group.name);
                                                                        closeDropdown();
                                                                    }}
                                                                    className="block w-full text-left px-4 py-2 text-sm text-red-600 hover:bg-gray-100"
                                                                >
                                                                    <Trash2 className="w-4 h-4 inline mr-2" />
                                                                    Delete
                                                                </button>
                                                            </div>
                                                        </div>
                                                    </>
                                                )}
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            ))}
                        </div>
                    )}
                </div>
            </WithTabbedUserLayout>
        </WithAdminBodyLayout>
    );
}

interface CreateGroupFormProps {
    onSubmit: (data: { name: string; info: string }) => Promise<void>;
}

function CreateGroupForm({ onSubmit }: CreateGroupFormProps) {
    const [name, setName] = useState('');
    const [info, setInfo] = useState('');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const { modal } = useGApp();

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!name.trim()) return;

        setLoading(true);
        setError('');
        try {
            await onSubmit({ name: name.trim(), info: info.trim() });
        } catch (err: any) {
            setError(err?.response?.data?.message || err?.message || 'An error occurred');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="space-y-4 md:w-md">
            <form onSubmit={handleSubmit} className="space-y-4">
                <div>
                    <label htmlFor="name" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                        Group Name *
                    </label>
                    <input
                        type="text"
                        id="name"
                        value={name}
                        onChange={(e) => setName(e.target.value)}
                        className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:text-white"
                        placeholder="Enter group name"
                        required
                    />
                </div>
                <div>
                    <label htmlFor="info" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                        Description
                    </label>
                    <textarea
                        id="info"
                        value={info}
                        onChange={(e) => setInfo(e.target.value)}
                        className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:text-white"
                        placeholder="Enter group description (optional)"
                        rows={3}
                    />
                </div>

                {error && (
                    <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-3">
                        <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
                    </div>
                )}

                <div className="flex gap-2 justify-end">
                    <button
                        type="button"
                        onClick={() => modal.closeModal()}
                        className="bg-gray-500 hover:bg-gray-600 text-white px-4 py-2 rounded-lg transition-colors"
                        disabled={loading}
                    >
                        Cancel
                    </button>
                    <button
                        type="submit"
                        className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg transition-colors disabled:opacity-50"
                        disabled={loading || !name.trim()}
                    >
                        {loading ? 'Creating...' : 'Create Group'}
                    </button>
                </div>
            </form>
        </div>
    );
}

interface EditGroupFormProps {
    group: UserGroup;
    onSubmit: (name: string, data: { info: string }) => Promise<void>;
}

function EditGroupForm({ group, onSubmit }: EditGroupFormProps) {
    const [info, setInfo] = useState(group.info || '');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const { modal } = useGApp();

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);
        setError('');
        try {
            await onSubmit(group.name, { info: info.trim() });
        } catch (err: any) {
            setError(err?.response?.data?.message || err?.message || 'An error occurred');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="space-y-4 md:w-md">
            <form onSubmit={handleSubmit} className="space-y-4">
                <div>
                    <label htmlFor="name" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                        Group Name
                    </label>
                    <input
                        type="text"
                        id="name"
                        value={group.name}
                        disabled
                        className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm bg-gray-100 dark:bg-gray-600 text-gray-500 dark:text-gray-400"
                    />
                </div>
                <div>
                    <label htmlFor="info" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                        Description
                    </label>
                    <textarea
                        id="info"
                        value={info}
                        onChange={(e) => setInfo(e.target.value)}
                        className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:text-white"
                        placeholder="Enter group description (optional)"
                        rows={3}
                    />
                </div>

                {error && (
                    <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-3">
                        <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
                    </div>
                )}

                <div className="flex gap-2 justify-end">
                    <button
                        type="button"
                        onClick={() => modal.closeModal()}
                        className="bg-gray-500 hover:bg-gray-600 text-white px-4 py-2 rounded-lg transition-colors"
                        disabled={loading}
                    >
                        Cancel
                    </button>
                    <button
                        type="submit"
                        className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg transition-colors disabled:opacity-50"
                        disabled={loading}
                    >
                        {loading ? 'Updating...' : 'Update Group'}
                    </button>
                </div>
            </form>
        </div>
    );
}
