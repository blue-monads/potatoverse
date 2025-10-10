"use client"

import { useState } from "react";
import { useGApp } from "@/hooks";
import { createUserInvite, UserInviteResponse } from "@/lib/api";
import { Copy, Check } from "lucide-react";

interface AddInviteModalProps {
    onInviteAdded: () => void;
}

export default function AddInviteModal({ onInviteAdded }: AddInviteModalProps) {
    const { modal } = useGApp();
    const [formData, setFormData] = useState({
        email: "",
        role: "",
        invited_as_type: "normal"
    });
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState("");
    const [createdInvite, setCreatedInvite] = useState<UserInviteResponse | null>(null);
    const [copied, setCopied] = useState(false);

    const invitedAsTypeOptions = [
        { value: "admin", label: "Admin" },
        { value: "moderator", label: "Moderator" },
        { value: "normal", label: "Normal User" },
        { value: "developer", label: "Developer" }
    ];

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setIsLoading(true);
        setError("");

        try {
            const response = await createUserInvite(formData);
            setCreatedInvite(response.data);
            onInviteAdded();
        } catch (err: any) {
            setError(err?.response?.data?.message || err?.message || 'An error occurred');
        } finally {
            setIsLoading(false);
        }
    };

    const handleInputChange = (field: string, value: string) => {
        setFormData(prev => ({
            ...prev,
            [field]: value
        }));
    };

    const handleCopyUrl = async () => {
        if (createdInvite?.invite_url) {
            try {
                await navigator.clipboard.writeText(createdInvite.invite_url);
                setCopied(true);
                setTimeout(() => setCopied(false), 2000);
            } catch (err) {
                console.error('Failed to copy URL:', err);
            }
        }
    };

    const handleClose = () => {
        setCreatedInvite(null);
        setCopied(false);
        modal.closeModal();
    };

    // Show success state with invite URL
    if (createdInvite) {
        return (
            <div className="space-y-4 md:w-md">
                <h2 className="text-xl font-semibold mb-4 text-green-600">✓ Invite Created Successfully</h2>
                
                <div className="bg-green-50 border border-green-200 rounded-lg p-4">
                    <p className="text-sm text-green-800 mb-3">
                        An invite has been sent to <strong>{createdInvite.email}</strong> and they can also use the direct link below:
                    </p>
                    
                    <div className="bg-white border border-green-300 rounded-md p-3">
                        <div className="flex items-center justify-between">
                            <div className="flex-1 min-w-0">
                                <p className="text-xs text-gray-500 mb-1">Invite URL:</p>
                                <p className="text-sm text-gray-900 break-all">
                                    {createdInvite.invite_url}
                                </p>
                            </div>
                            <button
                                onClick={handleCopyUrl}
                                className="ml-3 p-2 text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded-md transition-colors"
                                title="Copy URL"
                            >
                                {copied ? (
                                    <Check className="w-4 h-4 text-green-600" />
                                ) : (
                                    <Copy className="w-4 h-4" />
                                )}
                            </button>
                        </div>
                    </div>
                    
                    {copied && (
                        <p className="text-xs text-green-600 mt-2">✓ URL copied to clipboard!</p>
                    )}
                </div>

                <div className="flex justify-end space-x-2 pt-4">
                    <button
                        type="button"
                        onClick={handleClose}
                        className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
                    >
                        Done
                    </button>
                </div>
            </div>
        );
    }

    // Show form for creating invite
    return (
        <div className="space-y-4 md:w-md">
            <h2 className="text-xl font-semibold mb-4">Invite User</h2>
            
            <form onSubmit={handleSubmit} className="space-y-4">
                <div>
                    <label className="block text-sm font-medium mb-2">
                        Email Address
                    </label>
                    <input
                        type="email"
                        value={formData.email}
                        onChange={(e) => handleInputChange('email', e.target.value)}
                        placeholder="user@example.com"
                        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                        required
                    />
                </div>

                <div>
                    <label className="block text-sm font-medium mb-2">
                        Invite As Type
                    </label>
                    <select
                        value={formData.invited_as_type}
                        onChange={(e) => handleInputChange('invited_as_type', e.target.value)}
                        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                        required
                    >
                        {invitedAsTypeOptions.map((option) => (
                            <option key={option.value} value={option.value}>
                                {option.label}
                            </option>
                        ))}
                    </select>
                </div>

                {error && (
                    <div className="text-red-600 text-sm">
                        {error}
                    </div>
                )}

                <div className="flex justify-end space-x-2 pt-4">
                    <button
                        type="button"
                        onClick={modal.closeModal}
                        disabled={isLoading}
                        className="px-4 py-2 border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50"
                    >
                        Cancel
                    </button>
                    <button
                        type="submit"
                        disabled={isLoading}
                        className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
                    >
                        {isLoading ? "Creating..." : "Create Invite"}
                    </button>
                </div>
            </form>
        </div>
    );
}
