"use client";
import { Loader2 } from 'lucide-react';
import { useSearchParams } from 'next/navigation';
import React, { useEffect, useRef, useState } from 'react';

// /portal/admin/exec?nskey=test



export default function Page() {
    const searchParams = useSearchParams();
    const nskey = searchParams.get('nskey');
    const [isLoading, setIsLoading] = useState(true);

    useEffect(() => {
        setTimeout(() => {
            setIsLoading(false);
        }, 300);

    }, []);



    return (<>

        <div className='p-1'>

            <div className='p-1 rounded-md w-full min-h-[99vh] border  border-primary-100'>
                {isLoading && (
                    <div className='flex items-center justify-center h-full'>
                        <Loader2 className='w-12 h-12 animate-spin my-20' />
                    </div>
                )}

                {!isLoading && (
                    <iframe src={`/z/space/${nskey}`} className='w-full h-full'>
                    </iframe>
                )}
            </div>
            

        </div>

    </>)
}



