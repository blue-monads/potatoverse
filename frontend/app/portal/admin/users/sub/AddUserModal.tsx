"use client";
import React, { useState } from 'react';
import { createUserDirectly } from '@/lib';
import { useGApp } from '@/hooks';

interface AddUserModalProps {
    onUserAdded: () => void;
}

export default function AddUserModal({ onUserAdded }: AddUserModalProps) {
    const { modal } = useGApp();
    const [formData, setFormData] = useState({
        name: "",
        email: "",
        username: "",
        utype: "normal"
    });
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState("");
    const [createdUser, setCreatedUser] = useState<any>(null);

    const utypeOptions = [
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
            const response = await createUserDirectly(formData);
            setCreatedUser(response.data);
            // Don't close modal yet, show the created user info
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

    const handleClose = () => {
        if (createdUser) {
            onUserAdded();
        }
        modal.closeModal();
    };

    if (createdUser) {
        return (
            <div className="space-y-4 md:w-md">
                <div className="text-center">
                    <div className="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-green-100 mb-4">
                        <svg className="h-6 w-6 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                        </svg>
                    </div>
                    <h3 className="text-lg font-medium text-gray-900 dark:text-white">User Created Successfully!</h3>
                    <p className="text-sm text-gray-500 dark:text-gray-400 mt-2">
                        The user has been created and can now log in with the generated password.
                    </p>
                </div>

                <div className="bg-gray-50 dark:bg-gray-700 rounded-lg p-4 space-y-3">
                    <div>
                        <label className="text-sm font-medium text-gray-700 dark:text-gray-300">Name</label>
                        <p className="text-sm text-gray-900 dark:text-white">{createdUser.name}</p>
                    </div>
                    <div>
                        <label className="text-sm font-medium text-gray-700 dark:text-gray-300">Email</label>
                        <p className="text-sm text-gray-900 dark:text-white">{createdUser.email}</p>
                    </div>
                    <div>
                        <label className="text-sm font-medium text-gray-700 dark:text-gray-300">Username</label>
                        <p className="text-sm text-gray-900 dark:text-white">{createdUser.username}</p>
                    </div>
                    <div>
                        <label className="text-sm font-medium text-gray-700 dark:text-gray-300">Type</label>
                        <p className="text-sm text-gray-900 dark:text-white capitalize">{createdUser.utype}</p>
                    </div>
                    <div>
                        <label className="text-sm font-medium text-gray-700 dark:text-gray-300">Generated Password</label>
                        <div className="flex items-center space-x-2">
                            <p className="text-sm font-mono text-gray-900 dark:text-white bg-gray-100 dark:bg-gray-600 px-2 py-1 rounded">
                                {createdUser.password}
                            </p>
                            <button
                                type="button"
                                onClick={() => navigator.clipboard.writeText(createdUser.password)}
                                className="text-blue-600 hover:text-blue-800 text-sm"
                            >
                                Copy
                            </button>
                        </div>
                        <p className="text-xs text-red-600 dark:text-red-400 mt-1">
                            ⚠️ Please save this password securely. It won't be shown again.
                        </p>
                    </div>
                </div>

                <div className="flex gap-2 justify-end">
                    <button
                        onClick={handleClose}
                        className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg transition-colors"
                    >
                        Done
                    </button>
                </div>
            </div>
        );
    }

    return (
        <div className="space-y-4 md:w-md">
            <div>
                <h3 className="text-lg font-medium text-gray-900 dark:text-white">Create New User</h3>
                <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
                    Create a new user account directly. A random password will be generated.
                </p>
            </div>

            <form onSubmit={handleSubmit} className="space-y-4">
                <div>
                    <label htmlFor="name" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                        Full Name *
                    </label>
                    <input
                        type="text"
                        id="name"
                        value={formData.name}
                        onChange={(e) => handleInputChange("name", e.target.value)}
                        className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:text-white"
                        placeholder="Enter full name"
                        required
                    />
                </div>

                <div>
                    <label htmlFor="email" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                        Email Address *
                    </label>
                    <input
                        type="email"
                        id="email"
                        value={formData.email}
                        onChange={(e) => handleInputChange("email", e.target.value)}
                        className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:text-white"
                        placeholder="Enter email address"
                        required
                    />
                </div>

                <div>
                    <label htmlFor="username" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                        Username *
                    </label>
                    <input
                        type="text"
                        id="username"
                        value={formData.username}
                        onChange={(e) => handleInputChange("username", e.target.value)}
                        className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:text-white"
                        placeholder="Enter username"
                        required
                    />
                </div>

                <div>
                    <label htmlFor="utype" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                        User Type *
                    </label>
                    <select
                        id="utype"
                        value={formData.utype}
                        onChange={(e) => handleInputChange("utype", e.target.value)}
                        className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:text-white"
                        required
                    >
                        {utypeOptions.map((option) => (
                            <option key={option.value} value={option.value}>
                                {option.label}
                            </option>
                        ))}
                    </select>
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
                        disabled={isLoading}
                    >
                        Cancel
                    </button>
                    <button
                        type="submit"
                        className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg transition-colors disabled:opacity-50"
                        disabled={isLoading}
                    >
                        {isLoading ? 'Creating...' : 'Create User'}
                    </button>
                </div>
            </form>
        </div>
    );
}