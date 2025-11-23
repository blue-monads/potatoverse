"use client";
import React, { useEffect, useState } from 'react';
import { BookOpen, Code, FileText, ChevronRight } from 'lucide-react';
import { getDocsFile, getDocsIndex } from '@/lib/api';
import { Marked } from 'marked';
import './markdown.css';

const ICONS = {
    "BookOpen": BookOpen,
    "Code": Code,
    "FileText": FileText,
    "ChevronRight": ChevronRight,
} as const;

type DocSection = 'api-docs' | 'topics' | 'bindings-docs';

export default function Page() {
    const [mode, setMode] = useState<'root-page' | 'markdown-page'>('root-page');

    const [activeSection, setActiveSection] = useState<DocSection>('api-docs');
    const [expandedSection, setExpandedSection] = useState<DocSection | null>('api-docs');
    const [sections, setSections] = useState<any[]>([]);
    const [markdownPageUrl, setMarkdownPageUrl] = useState<string | null>(null);

    const loadIndex = async () => {
        try {
            const resp = await getDocsIndex();
            setSections(resp.data);
        } catch (error) {
            console.error("@error/1", error);
        }

    }

    useEffect(() => {

        loadIndex();

    }, []);



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
                        const Icon = ICONS[section.icon as keyof typeof ICONS];
                        const isActive = activeSection === section.id;
                        const isExpanded = expandedSection === section.id;

                        return (
                            <div key={section.id} className="mb-1">
                                <button
                                    onClick={() => {
                                        setActiveSection(section.id);
                                        setExpandedSection(isExpanded ? null : section.id);
                                    }}
                                    className={`w-full flex items-center justify-between px-3 py-2 rounded-lg text-sm font-medium transition-colors ${isActive
                                        ? 'bg-blue-100 text-blue-700'
                                        : 'text-gray-700 hover:bg-gray-100'
                                        }`}
                                >
                                    <div className="flex items-center gap-2">
                                        {Icon && <Icon className="w-4 h-4" />}
                                        <span>{section.title}</span>
                                    </div>
                                    <ChevronRight
                                        className={`w-4 h-4 transition-transform ${isExpanded ? 'rotate-90' : ''}`}
                                    />
                                </button>
                                {isExpanded && (
                                    <div className="ml-6">
                                        <ul className="list-disc list-inside">
                                            {section.content.items.map((item: any, index: number) => (
                                                <li
                                                    key={index}
                                                    className="text-gray-600 text-sm py-1 cursor-pointer hover:text-blue-600"
                                                    onClick={() => {

                                                        console.log("@item", item);

                                                        if (item.url) {
                                                            setMarkdownPageUrl(item.url);
                                                            setMode('markdown-page');
                                                        }
                                                    }}
                                                >
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

                {mode === 'root-page' && (

                    <>
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
                                    {currentSection.content.items.map((item: any, index: number) => (
                                        <div
                                            key={index}
                                            onClick={() => {
                                                if (item.url) {
                                                    setMarkdownPageUrl(item.url);
                                                    setMode('markdown-page');
                                                }
                                            }}
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


                            </div>
                        )}

                    </>

                )}




                {markdownPageUrl && (
                    <MarkdownPage url={markdownPageUrl} />
                )}
            </div>
        </div>
    );
}

interface MarkdownPageProps {
    url: string;
}

const MarkdownPage = (props: MarkdownPageProps) => {
    const [markdown, setMarkdown] = useState<string | null>(null);

    useEffect(() => {
        const fetchMarkdown = async () => {
            const resp = await getDocsFile(props.url);
            const instance = new Marked();
            const html = await instance.parse(resp.data);

            if (!html) {
                return;
            }

            setMarkdown(html as string);
        }
        fetchMarkdown();
    }, [props.url]);

    return (<>

        <div className="markdown-body">

            {markdown && (
                <div dangerouslySetInnerHTML={{ __html: markdown }} />
            )}

        </div>




    </>);

}