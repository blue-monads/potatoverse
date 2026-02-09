"use client";
import React, { useEffect, useRef, useState } from 'react';


import { Terminal } from 'lucide-react';
import WithAdminBodyLayout from '@/contain/Layouts/WithAdminBodyLayout';

export default function Page() {
    return (
        <WithAdminBodyLayout
            Icon={Terminal}
            name="Logs"
            description="View package logs"
            variant="none"
        >
            <div className="p-6">
                <div>Logs</div>
            </div>
        </WithAdminBodyLayout>
    );
}


