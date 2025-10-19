"use client";
import React, { useState, useEffect } from 'react';
import { Upload, CheckCircle, XCircle, FileIcon, AlertCircle, Copy } from 'lucide-react';
import { useSearchParams } from 'next/navigation';
import { uploadFileWithPresignedToken } from '@/lib';

export default function PresignedUploadPage() {
    const searchParams = useSearchParams();
    const [presignedKey, setPresignedKey] = useState('');
    const [file, setFile] = useState<File | null>(null);
    const [uploading, setUploading] = useState(false);
    const [uploadResult, setUploadResult] = useState<{ success: boolean; message: string; fileId?: number } | null>(null);
    const [error, setError] = useState('');

    useEffect(() => {
        const keyFromUrl = searchParams.get('presigned-key');
        if (keyFromUrl) {
            setPresignedKey(keyFromUrl);
        }
    }, [searchParams]);

    const handleFileSelect = (event: React.ChangeEvent<HTMLInputElement>) => {
        const selectedFile = event.target.files?.[0];
        if (selectedFile) {
            setFile(selectedFile);
            setError('');
            setUploadResult(null);
        }
    };

    const handleUpload = async () => {
        if (!presignedKey.trim()) {
            setError('Please enter a presigned key');
            return;
        }

        if (!file) {
            setError('Please select a file to upload');
            return;
        }

        setUploading(true);
        setError('');
        setUploadResult(null);

        try {
            const response = await uploadFileWithPresignedToken(presignedKey, file);
            setUploadResult({
                success: true,
                message: response.data.message || 'File uploaded successfully!',
                fileId: response.data.file_id
            });
            setFile(null);
            // Reset file input
            const fileInput = document.getElementById('file-input') as HTMLInputElement;
            if (fileInput) fileInput.value = '';
        } catch (err: any) {
            console.error('Upload error:', err);
            const errorMessage = err.response?.data?.error || err.message || 'Upload failed. Please check your token and try again.';
            setUploadResult({
                success: false,
                message: errorMessage
            });
        } finally {
            setUploading(false);
        }
    };

    const copyCurrentUrl = () => {
        const url = window.location.origin + window.location.pathname + (presignedKey ? `?presigned-key=${presignedKey}` : '');
        navigator.clipboard.writeText(url);
    };

    return (
        <div className="min-h-screen bg-gradient-to-br from-blue-50 via-white to-purple-50">
            <div className="container mx-auto px-4 py-12">
                <div className="max-w-2xl mx-auto">
                    {/* Header */}
                    <div className="text-center mb-8">
                        <div className="inline-flex items-center justify-center w-16 h-16 bg-blue-100 rounded-full mb-4">
                            <Upload className="w-8 h-8 text-blue-600" />
                        </div>
                        <h1 className="text-3xl font-bold text-gray-900 mb-2">
                            Presigned File Upload
                        </h1>
                        <p className="text-gray-600">
                            Upload files using a presigned token - no authentication required
                        </p>
                    </div>

                    {/* Main Card */}
                    <div className="bg-white rounded-xl shadow-lg p-8 space-y-6">
                        {/* Info Banner */}
                        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 flex items-start gap-3">
                            <AlertCircle className="w-5 h-5 text-blue-600 flex-shrink-0 mt-0.5" />
                            <div className="text-sm text-blue-900">
                                <p className="font-semibold mb-1">How it works:</p>
                                <ol className="list-decimal list-inside space-y-1 text-blue-800">
                                    <li>Paste your presigned token below</li>
                                    <li>Select the file to upload (must match the token's filename)</li>
                                    <li>Click "Upload File"</li>
                                </ol>
                            </div>
                        </div>

                        {/* Presigned Key Input */}
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-2">
                                Presigned Token
                            </label>
                            <div className="relative">
                                <textarea
                                    value={presignedKey}
                                    onChange={(e) => setPresignedKey(e.target.value)}
                                    placeholder="Paste your presigned token here..."
                                    rows={3}
                                    className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 font-mono text-sm resize-none"
                                />
                                {presignedKey && (
                                    <button
                                        onClick={copyCurrentUrl}
                                        className="absolute top-2 right-2 p-2 text-gray-400 hover:text-gray-600 bg-white rounded"
                                        title="Copy shareable URL"
                                    >
                                        <Copy className="w-4 h-4" />
                                    </button>
                                )}
                            </div>
                            {presignedKey && (
                                <p className="text-xs text-gray-500 mt-1">
                                    Token detected ({presignedKey.length} characters)
                                </p>
                            )}
                        </div>

                        {/* File Upload */}
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-2">
                                Select File
                            </label>
                            <div className="border-2 border-dashed border-gray-300 rounded-lg p-8 text-center hover:border-blue-400 transition-colors">
                                {file ? (
                                    <div className="space-y-3">
                                        <FileIcon className="w-12 h-12 text-blue-500 mx-auto" />
                                        <div>
                                            <p className="font-medium text-gray-900">{file.name}</p>
                                            <p className="text-sm text-gray-500">
                                                {(file.size / 1024 / 1024).toFixed(2)} MB
                                            </p>
                                        </div>
                                        <button
                                            onClick={() => {
                                                setFile(null);
                                                const fileInput = document.getElementById('file-input') as HTMLInputElement;
                                                if (fileInput) fileInput.value = '';
                                            }}
                                            className="text-sm text-red-600 hover:text-red-800"
                                        >
                                            Remove file
                                        </button>
                                    </div>
                                ) : (
                                    <div>
                                        <Upload className="w-12 h-12 text-gray-400 mx-auto mb-3" />
                                        <p className="text-gray-600 mb-2">
                                            Click to select a file
                                        </p>
                                        <p className="text-xs text-gray-500">
                                            File must match the token's expected filename
                                        </p>
                                    </div>
                                )}
                                <input
                                    id="file-input"
                                    type="file"
                                    onChange={handleFileSelect}
                                    className="hidden"
                                />
                                {!file && (
                                    <button
                                        onClick={() => document.getElementById('file-input')?.click()}
                                        className="mt-4 px-6 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200"
                                    >
                                        Choose File
                                    </button>
                                )}
                            </div>
                        </div>

                        {/* Error Message */}
                        {error && (
                            <div className="bg-red-50 border border-red-200 rounded-lg p-4 flex items-start gap-3">
                                <XCircle className="w-5 h-5 text-red-600 flex-shrink-0" />
                                <p className="text-sm text-red-800">{error}</p>
                            </div>
                        )}

                        {/* Upload Result */}
                        {uploadResult && (
                            <div className={`border rounded-lg p-4 flex items-start gap-3 ${
                                uploadResult.success 
                                    ? 'bg-green-50 border-green-200' 
                                    : 'bg-red-50 border-red-200'
                            }`}>
                                {uploadResult.success ? (
                                    <CheckCircle className="w-5 h-5 text-green-600 flex-shrink-0" />
                                ) : (
                                    <XCircle className="w-5 h-5 text-red-600 flex-shrink-0" />
                                )}
                                <div className="flex-1">
                                    <p className={`text-sm font-medium ${
                                        uploadResult.success ? 'text-green-900' : 'text-red-900'
                                    }`}>
                                        {uploadResult.message}
                                    </p>
                                    {uploadResult.success && uploadResult.fileId && (
                                        <p className="text-xs text-green-700 mt-1">
                                            File ID: {uploadResult.fileId}
                                        </p>
                                    )}
                                </div>
                            </div>
                        )}

                        {/* Upload Button */}
                        <button
                            onClick={handleUpload}
                            disabled={uploading || !presignedKey.trim() || !file}
                            className="w-full px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed font-medium transition-colors flex items-center justify-center gap-2"
                        >
                            {uploading ? (
                                <>
                                    <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-white"></div>
                                    Uploading...
                                </>
                            ) : (
                                <>
                                    <Upload className="w-5 h-5" />
                                    Upload File
                                </>
                            )}
                        </button>
                    </div>

                    {/* Help Section */}
                    <div className="mt-8 bg-gray-50 rounded-lg p-6">
                        <h3 className="font-semibold text-gray-900 mb-3">Need Help?</h3>
                        <div className="space-y-2 text-sm text-gray-600">
                            <p>
                                <strong>Where do I get a presigned token?</strong><br />
                                Tokens are generated by space owners in the file management interface. Ask your administrator for a token.
                            </p>
                            <p>
                                <strong>Can I share this page?</strong><br />
                                Yes! Once you paste a token, click the copy icon to get a shareable URL with the token included.
                            </p>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}