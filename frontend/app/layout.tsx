"use client"
import "./globals.css";
import React from 'react';
import { Search, Home, Grid, FileText, User, Settings, Plus, Zap } from 'lucide-react';

import { Suspense } from "react";
import { GAppStateContext } from "@/hooks/contexts/GAppStateContext";
import GModalWrapper from "@/hooks/modal/GModalWrapper";
import { staticGradients } from "./utils";



export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      data-theme="cerberus"
    >
      <body
        className={`antialiased text-gray-900 bg-surface-50`}

        // suppressHydrationWarning
      >
        <script src="/zz/api/core/global.js" />

        <div className="hidden">
          {staticGradients.map((gradient, index) => (
            <div key={index} className={`bg-gradient-to-br ${gradient} h-0 w-0 absolute`} />
          ))}
        </div>

        <GAppStateContext>
          <Suspense fallback={<div className="w-full h-full bg-gray-200 rounded animate-pulse"></div>}>
            {children}
          </Suspense>
          <GModalWrapper />

        </GAppStateContext>
      </body>
    </html>
  );
}



