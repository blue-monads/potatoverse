"use client"
import { AddButton } from "@/contain/AddButton";
import WithAdminBodyLayout from "@/contain/Layouts/WithAdminBodyLayout";
import { Mail, UserIcon, Clock, CheckCircle, XCircle, MoreVertical } from "lucide-react";
import WithTabbedUserLayout from "../WithTabbedUserLayout";
import BigSearchBar from "@/contain/compo/BigSearchBar";
import { useState, useEffect } from "react";
import { useGApp } from "@/hooks";
import AddInviteModal from "../sub/AddInviteModal";
import { getUserInvites, deleteUserInvite, resendUserInvite, UserInvite } from "@/lib/api";
// Using standard HTML elements instead of skeleton components

export default function Page() {
    const [searchTerm, setSearchTerm] = useState('');
    const [invites, setInvites] = useState<UserInvite[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [openDropdownId, setOpenDropdownId] = useState<number | null>(null);
    const gapp = useGApp();
    const modal = gapp.modal;

    const fetchInvites = async () => {
        try {
            const response = await getUserInvites();
            setInvites(response.data);
        } catch (error) {
            console.error('Failed to fetch invites:', error);
        } finally {
            setIsLoading(false);
        }
    };

    useEffect(() => {
        if (!gapp.isInitialized) {
            return;
        }

        fetchInvites();
    }, [gapp.isInitialized]);

    const getStatusIcon = (status: string) => {
        switch (status) {
            case 'pending':
                return <Clock className="w-4 h-4 text-yellow-500" />;
            case 'accepted':
                return <CheckCircle className="w-4 h-4 text-green-500" />;
            case 'rejected':
                return <XCircle className="w-4 h-4 text-red-500" />;
            default:
                return <Clock className="w-4 h-4 text-gray-500" />;
        }
    };

    const getStatusBadge = (status: string) => {
        const baseClasses = "px-2 py-1 text-xs font-medium rounded-full";
        switch (status) {
            case 'pending':
                return <span className={`${baseClasses} bg-yellow-100 text-yellow-800`}>Pending</span>;
            case 'accepted':
                return <span className={`${baseClasses} bg-green-100 text-green-800`}>Accepted</span>;
            case 'rejected':
                return <span className={`${baseClasses} bg-red-100 text-red-800`}>Rejected</span>;
            default:
                return <span className={`${baseClasses} bg-gray-100 text-gray-800`}>Unknown</span>;
        }
    };

    const handleResendInvite = async (id: number) => {
        try {
            await resendUserInvite(id);
            fetchInvites(); // Refresh the list
        } catch (error) {
            console.error('Failed to resend invite:', error);
        }
    };

    const handleDeleteInvite = async (id: number) => {
        if (!confirm('Are you sure you want to delete this invite?')) return;
        
        try {
            await deleteUserInvite(id);
            fetchInvites(); // Refresh the list
        } catch (error) {
            console.error('Failed to delete invite:', error);
        }
    };

    const toggleDropdown = (id: number) => {
        setOpenDropdownId(openDropdownId === id ? null : id);
    };

    const closeDropdown = () => {
        setOpenDropdownId(null);
    };

    const filteredInvites = invites.filter(invite =>
        invite.email.toLowerCase().includes(searchTerm.toLowerCase()) ||
        invite.role.toLowerCase().includes(searchTerm.toLowerCase()) ||
        invite.status.toLowerCase().includes(searchTerm.toLowerCase())
    );

    return (<>
        <WithAdminBodyLayout
            Icon={UserIcon}
            name='Users'
            description="Manage your users, roles, and permissions."
            rightContent={<>
                <AddButton
                    name="+ Invite User"
                    onClick={() => {
                        modal.openModal({
                            title: "Invite User",
                            content: <AddInviteModal onInviteAdded={fetchInvites} />
                        });
                    }}
                />
            </>}
        >
            <BigSearchBar
                setSearchText={setSearchTerm}
                searchText={searchTerm}
                placeholder="Search invites..."
            />

            <WithTabbedUserLayout activeTab="invites">
                <div className="max-w-7xl mx-auto">
                    {isLoading ? (
                        <div className="flex justify-center items-center h-32">
                            <div className="text-gray-500">Loading invites...</div>
                        </div>
                    ) : filteredInvites.length === 0 ? (
                        <div className="text-center py-12">
                            <Mail className="w-12 h-12 text-gray-400 mx-auto mb-4" />
                            <h3 className="text-lg font-medium text-gray-900 mb-2">No invites yet</h3>
                            <p className="text-gray-500 mb-4">Get started by inviting users to your platform.</p>
                            <button
                                className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
                                onClick={() => modal.openModal({
                                    title: "Invite User",
                                    content: <AddInviteModal onInviteAdded={fetchInvites} />
                                })}
                            >
                                Invite User
                            </button>
                        </div>
                    ) : (
                        <div className="space-y-4">
                            {filteredInvites.map((invite) => (
                                <div key={invite.id} className="bg-white border border-gray-200 rounded-lg p-4 shadow-sm">
                                    <div className="flex items-center justify-between">
                                        <div className="flex items-center space-x-4">
                                            <div className="flex items-center space-x-2">
                                                {getStatusIcon(invite.status)}
                                                <div>
                                                    <div className="font-medium">{invite.email}</div>
                                                    <div className="text-sm text-gray-500">
                                                        {invite.role} • {invite.invited_as_type}
                                                    </div>
                                                </div>
                                            </div>
                                        </div>
                                        <div className="flex items-center space-x-3">
                                            {getStatusBadge(invite.status)}
                                            <div className="text-sm text-gray-500">
                                                {new Date(invite.created_at).toLocaleDateString()}
                                            </div>
                                            <div className="relative">
                                                <button 
                                                    className="p-2 hover:bg-gray-100 rounded-md"
                                                    onClick={() => toggleDropdown(invite.id)}
                                                >
                                                    <MoreVertical className="w-4 h-4" />
                                                </button>
                                                {openDropdownId === invite.id && (
                                                    <>
                                                        <div 
                                                            className="fixed inset-0 z-10" 
                                                            onClick={closeDropdown}
                                                        ></div>
                                                        <div className="absolute right-0 mt-2 w-48 bg-white border border-gray-200 rounded-md shadow-lg z-20">
                                                            <div className="py-1">
                                                                {invite.status === 'pending' && (
                                                                    <button
                                                                        onClick={() => {
                                                                            handleResendInvite(invite.id);
                                                                            closeDropdown();
                                                                        }}
                                                                        className="block w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                                                                    >
                                                                        Resend Invite
                                                                    </button>
                                                                )}
                                                                <button
                                                                    onClick={() => {
                                                                        handleDeleteInvite(invite.id);
                                                                        closeDropdown();
                                                                    }}
                                                                    className="block w-full text-left px-4 py-2 text-sm text-red-600 hover:bg-gray-100"
                                                                >
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
    </>)
}