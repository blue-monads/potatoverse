"use client";
import { Loader2 } from 'lucide-react';
import { useSearchParams } from 'next/navigation';
import React, { useEffect, useRef, useState } from 'react';

// /portal/admin/exec?nskey=test&space_id=1

export default function Page() {
    const searchParams = useSearchParams();
    const nskey = searchParams.get('nskey');
    const space_id = searchParams.get('space_id');
    const [isLoading, setIsLoading] = useState(true);
    const iframeRef = useRef<HTMLIFrameElement>(null);
    const [iframeSrc, setIframeSrc] = useState('');

    useEffect(() => {
        if (!nskey || !space_id) {
            return;
        }


        const timer = setTimeout(() => {
            setIsLoading(false);
        }, 300);

        const attrs = (window as any).__potato_attrs__ || {};
        const host = attrs.site_host || '';
        if (host) {
            const origin = window.location.origin;
            const isSecure = origin.startsWith("https//");
            const hasPort = origin.includes(":");
            const port = hasPort ? origin.split(":").at(-1) : "";

            if (host.includes('*')) {
                setIframeSrc(`${isSecure ? "https://" : "http://"}${host.replace('*', "s-" + space_id) }${hasPort ? ":" + port : ""}/zz/space/${nskey}`);
            } else {
                setIframeSrc(`${isSecure ? "https://" : "http://"}${host}${hasPort ? ":" + port : ""}/zz/space/${nskey}`);
            }
        } else {
            setIframeSrc(`/zz/space/${nskey}`);
        }


        return () => clearTimeout(timer);
    }, [nskey, space_id]);

    console.log(iframeSrc);

    return (
        <div className='p-1'>
            <div className='p-1 rounded-md w-full min-h-[99vh] border border-primary-100 flex flex-col'>
                {isLoading ? (
                    <div className='flex items-center justify-center h-full'>
                        <Loader2 className='w-12 h-12 animate-spin my-20' />
                    </div>
                ) : (
                    <iframe
                        ref={iframeRef}
                        src={iframeSrc}
                        // src={`/zz/test_page.html`}
                        className='w-full h-full flex-grow'
                    ></iframe>
                )}
            </div>
        </div>
    );
}