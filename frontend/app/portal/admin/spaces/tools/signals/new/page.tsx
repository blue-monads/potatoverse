"use client";
import React from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import SignalEditor from '../components/SignalEditor';
import { createSignal } from '@/lib';

export default function Page() {
    const router = useRouter();
    const searchParams = useSearchParams();
    const installId = searchParams.get('install_id');
    const spaceId = searchParams.get('space_id');

    if (!installId) {
        return <div>Install ID not provided</div>;
    }

    const handleSave = async (data: any) => {
        try {
            await createSignal(parseInt(installId), data);
            const params = new URLSearchParams(searchParams.toString());
            router.push(`/portal/admin/spaces/tools/signals?${params.toString()}`);
        } catch (error) {
            console.error('Failed to create signal:', error);
            alert('Failed to create signal: ' + ((error as any)?.response?.data?.error || (error as any)?.message));
            throw error;
        }
    };

    const handleBack = () => {
        const params = new URLSearchParams(searchParams.toString());
        router.push(`/portal/admin/spaces/tools/signals?${params.toString()}`);
    };

    return (
        <SignalEditor
            onSave={handleSave}
            onBack={handleBack}
            initialData={null}
            installId={parseInt(installId)}
            spaceId={spaceId ? parseInt(spaceId) : undefined}
        />
    );
}
