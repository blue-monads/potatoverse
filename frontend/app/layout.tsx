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
        className={`antialiased`}
      >
        <script src="/zz/api/core/global.js" />

        <div className="hidden">
          {staticGradients.map((gradient, index) => (
            <div key={index} className={`bg-gradient-to-br ${gradient} h-0 w-0 absolute`} />
          ))}
        </div>

        <GAppStateContext>
          <Suspense fallback={<SkeletonLoader />}>
            {children}
          </Suspense>
          <GModalWrapper />



          {/* <SkeletonLoader /> */}
        </GAppStateContext>
      </body>
    </html>
  );
}



function SkeletonLoader() {
  return (
    <div className="flex h-screen bg-gray-50">
      {/* Sidebar */}
      <div className="w-14 bg-white border-r border-gray-200 flex flex-col items-center py-4 space-y-4">
        <div className="w-10 h-10 bg-orange-100 rounded-lg flex items-center justify-center">

        </div>
        <div className="w-8 h-8 bg-gray-200 rounded-lg animate-pulse"></div>
        <div className="w-8 h-8 bg-gray-200 rounded-lg animate-pulse"></div>
        <div className="w-8 h-8 bg-gray-200 rounded-lg animate-pulse"></div>
        <div className="w-8 h-8 bg-gray-200 rounded-lg animate-pulse"></div>

        <div className="flex-1"></div>
        <div className="w-8 h-8 bg-green-400 rounded-full"></div>
        <div className="w-8 h-8 bg-gray-200 rounded-lg animate-pulse"></div>
      </div>


      {/* Main Content */}
      <div className="flex-1 flex flex-col">
        <div className="bg-white border-b border-gray-200 py-6 flex items-center justify-between w-full">

          <div className="max-w-7xl mx-auto w-full px-8 flex items-center justify-between">
            <div className="flex items-center space-x-3 ">
              <div className="w-10 h-10 bg-purple-200 rounded-lg flex items-center justify-center">

              </div>
              <div>
                <div className="w-20 h-5 bg-gray-200 rounded animate-pulse mb-1"></div>
                <div className="w-32 h-3 bg-gray-200 rounded animate-pulse"></div>
              </div>
            </div>

            <div className="w-24 h-8 bg-gray-200 rounded animate-pulse"></div>
          </div>

        </div>


        <div className="flex-1 overflow-auto px-8 py-6">
          {/* Search Bar */}
          <div className="mb-8">
            <div className="relative">
              <Search className="absolute left-4 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" />
              <div className="w-full h-12 bg-white border border-gray-200 rounded-lg pl-12 animate-pulse"></div>
            </div>
          </div>

          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center space-x-2">

              <div className="w-32 h-5 bg-gray-200 rounded animate-pulse"></div>
            </div>
            <div className="w-32 h-6 bg-gray-200 rounded animate-pulse"></div>
          </div>


          {/* Cards */}
          <div className="flex flex-wrap gap-4">
            {Array.from({ length: 10 }).map((_, index) => (
              <div key={index} className="bg-gradient-to-br rounded-xl p-6 w-full max-w-sm relative overflow-hidden shadow">
                <div className="absolute inset-0 bg-gradient-to-r from-transparent via-white/10 to-transparent animate-pulse"></div>
                <div className="relative">
                  <div className="flex items-center space-x-2 mb-4 pb-10">
                    <div className="w-16 h-6 bg-purple-100/50 rounded-full animate-pulse"></div>
                  </div>
                </div>
              </div>
            ))}

          </div>


        </div>
      </div>
    </div>
  );
}