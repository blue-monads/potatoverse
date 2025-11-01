"use client";
import React from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import EventSubscriptionEditor from '../components/EventSubscriptionEditor';
import { createEventSubscription } from '@/lib';

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
            await createEventSubscription(parseInt(installId), {
                ...data,
                space_id: spaceId ? parseInt(spaceId) : undefined,
            });
            const params = new URLSearchParams();
            params.set('install_id', installId);
            if (spaceId) params.set('space_id', spaceId);
            router.push(`/portal/admin/spaces/tools/events?${params.toString()}`);
        } catch (error) {
            console.error('Failed to create event subscription:', error);
            alert('Failed to create event subscription: ' + ((error as any)?.response?.data?.error || (error as any)?.message));
            throw error;
        }
    };

    const handleBack = () => {
        const params = new URLSearchParams();
        params.set('install_id', installId);
        if (spaceId) params.set('space_id', spaceId);
        router.push(`/portal/admin/spaces/tools/events?${params.toString()}`);
    };

    return (
        <EventSubscriptionEditor
            onSave={handleSave}
            onBack={handleBack}
            initialData={null}
        />
    );
}

