"use client";
import React, { useState } from 'react';
import { BookOpen, Code, FileText, ChevronRight } from 'lucide-react';

type DocSection = 'api-docs' | 'topics' | 'bindings-docs';

export default function Page() {
    const [activeSection, setActiveSection] = useState<DocSection>('api-docs');
    const [expandedSection, setExpandedSection] = useState<DocSection | null>('api-docs');

    const sections = [
        {
            id: 'api-docs' as DocSection,
            title: 'API Docs',
            icon: Code,
            content: {
                title: 'API Documentation',
                description: 'Complete API reference for all available endpoints and methods.',
                items: [
                    { title: 'Space', description: 'Reference for all available endpoints and methods in the Space API.' },
                    { title: 'Plugin', description: 'Reference for all available endpoints and methods in the Plugin API.' },
                    { title: 'Capability', description: 'Reference for all available endpoints and methods in the Capability API.' },
                    { title: 'Event', description: 'Reference for all available endpoints and methods in the Event API.' },
                    { title: 'Auth ', description: 'Reference for all available endpoints and methods in the Auth API.' },
                    { title: 'User', description: 'Reference for all available endpoints and methods in the User API.' },
                ]
            }
        },
        {
            id: 'topics' as DocSection,
            title: 'Topics',
            icon: BookOpen,
            content: {
                title: 'Documentation Topics',
                description: 'Explore guides and tutorials on various topics.',
                items: [
                    { title: 'Getting Started', description: 'Introduction and quick start guide' },
                    { title: 'Core Concepts', description: 'Understanding the fundamental concepts' },
                    { title: 'Best Practices', description: 'Recommended patterns and practices' },
                    { title: 'Examples', description: 'Real-world examples and use cases' },
                ]
            }
        },
        {
            id: 'bindings-docs' as DocSection,
            title: 'Bindings Docs',
            icon: FileText,
            content: {
                title: 'Executors Bindings Documentation',
                description: 'Documentation for executors bindings.',
                items: [
                    { title: 'Luaz', description: 'Luaz bindings for the Luaz executor.' },
                ]
            }
        },
    ];

    const currentSection = sections.find(s => s.id === activeSection);

    return (
        <div className="flex h-[100vh] bg-white rounded-lg border border-gray-200">
            {/* Sidebar */}
            <div className="w-64 border-r border-gray-200 bg-gray-50 flex flex-col">
                <div className="p-4 border-b border-gray-200">
                    <h2 className="text-lg font-semibold text-gray-900 uppercase">Docs and Help</h2>
                </div>
                <nav className="flex-1 overflow-y-auto p-2">
                    {sections.map((section) => {
                        const Icon = section.icon;
                        const isActive = activeSection === section.id;
                        const isExpanded = expandedSection === section.id;
                        
                        return (
                            <div key={section.id} className="mb-1">
                                <button
                                    onClick={() => {
                                        setActiveSection(section.id);
                                        setExpandedSection(isExpanded ? null : section.id);
                                    }}
                                    className={`w-full flex items-center justify-between px-3 py-2 rounded-lg text-sm font-medium transition-colors ${
                                        isActive
                                            ? 'bg-blue-100 text-blue-700'
                                            : 'text-gray-700 hover:bg-gray-100'
                                    }`}
                                >
                                    <div className="flex items-center gap-2">
                                        <Icon className="w-4 h-4" />
                                        <span>{section.title}</span>
                                    </div>
                                    <ChevronRight 
                                        className={`w-4 h-4 transition-transform ${isExpanded ? 'rotate-90' : ''}`}
                                    />
                                </button>
                                {isExpanded && (
                                    <div className="ml-6">
                                        <ul className="list-disc list-inside">
                                                {section.content.items.map((item, index) => (
                                                <li key={index} className="text-gray-600 text-sm py-1">
                                                    {item.title}
                                                </li>
                                            ))}
                                        </ul>
                                    </div>
                                )}


                            </div>
                        );
                    })}
                </nav>
            </div>

            {/* Main Content */}
            <div className="flex-1 overflow-y-auto">
                {currentSection && (
                    <div className="p-8">
                        <div className="mb-6">
                            <h1 className="text-3xl font-bold text-gray-900 mb-2">
                                {currentSection.content.title}
                            </h1>
                            <p className="text-gray-600 text-lg">
                                {currentSection.content.description}
                            </p>
                        </div>

                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-8">
                            {currentSection.content.items.map((item, index) => (
                                <div
                                    key={index}
                                    className="p-6 border border-gray-200 rounded-lg hover:border-blue-300 hover:shadow-md transition-all cursor-pointer bg-white"
                                >
                                    <h3 className="text-xl font-semibold text-gray-900 mb-2">
                                        {item.title}
                                    </h3>
                                    <p className="text-gray-600">
                                        {item.description}
                                    </p>
                                </div>
                            ))}
                        </div>

                        {/* Placeholder content area */}
                        <div className="mt-8 p-6 bg-gray-50 rounded-lg border border-gray-200">
                            <p className="text-gray-500 italic">
                                Documentation content will be displayed here. This is a placeholder for the {currentSection.content.title.toLowerCase()} section.
                            </p>
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
}
