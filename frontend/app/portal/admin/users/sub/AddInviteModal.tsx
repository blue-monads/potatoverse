"use client"

import { useState } from "react";
import { useGApp } from "@/hooks";
import { createUserInvite } from "@/lib/api";

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
            await createUserInvite(formData);
            onInviteAdded();
            modal.closeModal();
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

    return (
        <div className="space-y-4">
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
