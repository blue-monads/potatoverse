"use client";
import { useGApp } from '@/hooks';
import useSimpleDataLoader from '@/hooks/useSimpleDataLoader';
import { getSpaceSpec } from '@/lib/api';
import { useSearchParams } from 'next/navigation';
import React, { useState, useMemo, useEffect, useRef } from 'react';
import {
    BookOpen,
    Zap,
    Box,
    Layers,
    Code,
    FileText,
    ChevronDown,
    ChevronRight,
    Copy,
    Check
} from 'lucide-react';
import WithAdminBodyLayout from '@/contain/Layouts/WithAdminBodyLayout';

type SpecType = 'models' | 'scopes' | 'events_outputs' | 'event_slots' | 'apis' | 'blocks';

interface SchemaField {
    type?: string;
    description?: string;
    properties?: Record<string, any>;
    items?: any;
    $ref?: string;
    [key: string]: any;
}

interface ModelSpec {
    name: string;
    description: string;
    schema: Record<string, any>;
}

interface ScopeSpec {
    name: string;
    description: string;
}

interface EventSpec {
    name: string;
    description: string;
    schema: SchemaField;
    schema_file?: string;
}

interface HandlerSpec {
    name: string;
    description: string;
    schema: SchemaField;
    schema_file?: string;
}

interface BlockSpec {
    name: string;
    description: string;
    schema: SchemaField;
    schema_file?: string;
}

interface SpaceSpec {
    scopes: ScopeSpec[];
    events_outputs: EventSpec[];
    event_slots: HandlerSpec[];
    APIs?: HandlerSpec[];
    blocks?: BlockSpec[];
}

interface PotatoSpec {
    space_specs: Record<string, SpaceSpec>;
    models: ModelSpec[];
}

function SchemaViewer({ schema, models }: { schema: SchemaField; models: ModelSpec[] }) {
    const [expanded, setExpanded] = useState(false);
    const [copied, setCopied] = useState(false);

    const handleCopy = () => {
        navigator.clipboard.writeText(JSON.stringify(schema, null, 2));
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    };

    const resolveRef = (ref: string): any => {
        if (ref?.startsWith('#/models/')) {
            const modelName = ref.replace('#/models/', '');
            const model = models.find(m => m.name === modelName);
            return model?.schema || {};
        }
        return {};
    };

    const renderSchemaValue = (value: any, key?: string, depth = 0): React.ReactNode => {
        if (depth > 3) return <span className="text-gray-400">...</span>;

        if (value === null) return <span className="text-gray-400">null</span>;
        if (value === undefined) return <span className="text-gray-400">undefined</span>;

        if (typeof value === 'string') {
            return <span className="text-green-600">"{value}"</span>;
        }
        if (typeof value === 'number' || typeof value === 'boolean') {
            return <span className="text-blue-600">{String(value)}</span>;
        }

        if (Array.isArray(value)) {
            return (
                <div className="ml-4">
                    <span className="text-gray-500">[</span>
                    {value.map((item, idx) => (
                        <div key={idx} className="ml-4">
                            {renderSchemaValue(item, undefined, depth + 1)}
                            {idx < value.length - 1 && <span className="text-gray-500">,</span>}
                        </div>
                    ))}
                    <span className="text-gray-500">]</span>
                </div>
            );
        }

        if (typeof value === 'object' && value !== null) {
            if (value.$ref) {
                const resolved = resolveRef(value.$ref);
                return (
                    <div className="ml-4">
                        <span className="text-purple-600">$ref: {value.$ref}</span>
                        {expanded && Object.keys(resolved).length > 0 && (
                            <div className="ml-4 mt-1 border-l-2 border-gray-200 pl-2">
                                {renderSchemaValue(resolved, undefined, depth + 1)}
                            </div>
                        )}
                    </div>
                );
            }

            const entries = Object.entries(value);
            if (entries.length === 0) return <span className="text-gray-400">{'{}'}</span>;

            return (
                <div className="ml-4">
                    <span className="text-gray-500">{'{'}</span>
                    {entries.map(([k, v], idx) => (
                        <div key={k} className="ml-4">
                            <span className="text-purple-600">{k}:</span>{' '}
                            {renderSchemaValue(v, k, depth + 1)}
                            {idx < entries.length - 1 && <span className="text-gray-500">,</span>}
                        </div>
                    ))}
                    <span className="text-gray-500">{'}'}</span>
                </div>
            );
        }

        return <span className="text-gray-600">{String(value)}</span>;
    };

    return (
        <div className="bg-gray-50 rounded-lg p-4 border border-gray-200">
            <div className="flex items-center justify-between mb-2">
                <button
                    onClick={() => setExpanded(!expanded)}
                    className="flex items-center gap-2 text-sm text-gray-600 hover:text-gray-900"
                >
                    {expanded ? (
                        <ChevronDown className="w-4 h-4" />
                    ) : (
                        <ChevronRight className="w-4 h-4" />
                    )}
                    <span className="font-medium">Schema</span>
                </button>
                <button
                    onClick={handleCopy}
                    className="flex items-center gap-1 text-xs text-gray-500 hover:text-gray-700"
                >
                    {copied ? (
                        <>
                            <Check className="w-3 h-3" />
                            <span>Copied!</span>
                        </>
                    ) : (
                        <>
                            <Copy className="w-3 h-3" />
                            <span>Copy JSON</span>
                        </>
                    )}
                </button>
            </div>
            {expanded && (
                <div className="mt-2 font-mono text-sm overflow-x-auto">
                    {renderSchemaValue(schema)}
                </div>
            )}
        </div>
    );
}

function SpecSection({ title, icon: Icon, children }: { title: string; icon: any; children: React.ReactNode }) {
    return (
        <div>
            <div className="flex items-center gap-3 mb-6">
                <Icon className="w-6 h-6 text-gray-600" />
                <h2 className="text-3xl font-bold text-gray-900">{title}</h2>
            </div>
            {children}
        </div>
    );
}

function SpecCard({ title, description, schema, models, schemaFile }: {
    title: string;
    description: string;
    schema?: SchemaField;
    models?: ModelSpec[];
    schemaFile?: string;
}) {
    const hasArgsResult = schema?.properties?.args || schema?.properties?.result;

    return (
        <div className="bg-white rounded-lg border border-gray-200 p-5 hover:shadow-md transition-shadow">
            <div className="flex items-start justify-between mb-3">
                <div className="flex-1">
                    <h3 className="text-lg font-semibold text-gray-900 mb-1">{title}</h3>
                    {description && (
                        <p className="text-sm text-gray-600 mb-3">{description}</p>
                    )}
                    {schemaFile && (
                        <div className="flex items-center gap-2 text-xs text-gray-500 mb-3">
                            <FileText className="w-3 h-3" />
                            <span>Schema file: {schemaFile}</span>
                        </div>
                    )}
                </div>
            </div>
            {schema && models && (
                <div className="space-y-3">
                    {hasArgsResult ? (
                        <>
                            {schema.properties?.args && (
                                <div>
                                    <div className="flex items-center gap-2 mb-2">
                                        <Code className="w-4 h-4 text-gray-500" />
                                        <span className="text-sm font-semibold text-gray-700">Arguments</span>
                                    </div>
                                    <SchemaViewer schema={schema.properties.args} models={models} />
                                </div>
                            )}
                            {schema.properties?.result && (
                                <div>
                                    <div className="flex items-center gap-2 mb-2">
                                        <Zap className="w-4 h-4 text-gray-500" />
                                        <span className="text-sm font-semibold text-gray-700">Result</span>
                                        {schema.properties.result.description && (
                                            <span className="text-xs text-gray-500 italic">
                                                - {schema.properties.result.description}
                                            </span>
                                        )}
                                    </div>
                                    <SchemaViewer schema={schema.properties.result} models={models} />
                                </div>
                            )}
                        </>
                    ) : (
                        <SchemaViewer schema={schema} models={models} />
                    )}
                </div>
            )}
        </div>
    );
}

export default function Page() {
    const searchParams = useSearchParams();
    const installId = searchParams.get('install_id');
    const gapp = useGApp();

    // Call all hooks unconditionally at the top
    const loader = useSimpleDataLoader<PotatoSpec>({
        loader: () => installId ? getSpaceSpec(parseInt(installId)) : Promise.resolve(null as any),
        ready: gapp.isInitialized && !!installId,
    });

    const spec = loader.data;
    const spaceSpecKeys = spec ? Object.keys(spec.space_specs || {}) : [];
    const firstSpaceKey = spaceSpecKeys[0];
    const spaceSpec = spec && firstSpaceKey ? spec.space_specs[firstSpaceKey] : null;

    // Available spec types with counts - always show all types, even if empty
    const availableTypes = useMemo(() => {
        const types: Array<{ type: SpecType; label: string; icon: any; count: number }> = [
            { 
                type: 'models', 
                label: 'Models', 
                icon: Layers, 
                count: spec?.models?.length || 0 
            },
            { 
                type: 'scopes', 
                label: 'Scopes', 
                icon: Box, 
                count: spaceSpec?.scopes?.length || 0 
            },
            { 
                type: 'events_outputs', 
                label: 'Event Outputs', 
                icon: Zap, 
                count: spaceSpec?.events_outputs?.length || 0 
            },
            { 
                type: 'event_slots', 
                label: 'Event Slots', 
                icon: Code, 
                count: spaceSpec?.event_slots?.length || 0 
            },
            { 
                type: 'apis', 
                label: 'APIs', 
                icon: Code, 
                count: spaceSpec?.APIs?.length || 0 
            },
            { 
                type: 'blocks', 
                label: 'Blocks', 
                icon: Box, 
                count: spaceSpec?.blocks?.length || 0 
            },
        ];
        
        return types;
    }, [spec, spaceSpec]);

    const [selectedType, setSelectedType] = useState<SpecType | null>(null);
    const initializedRef = useRef(false);

    // Update selectedType when availableTypes changes (only set initial value)
    useEffect(() => {
        if (availableTypes.length > 0) {
            if (!initializedRef.current) {
                // Select first type that has items, or first type if all are empty
                const firstTypeWithItems = availableTypes.find(t => t.count > 0);
                setSelectedType(firstTypeWithItems ? firstTypeWithItems.type : availableTypes[0].type);
                initializedRef.current = true;
            } else {
                // If current selection is no longer available, reset to first
                const currentTypeExists = availableTypes.find(t => t.type === selectedType);
                if (!currentTypeExists) {
                    const firstTypeWithItems = availableTypes.find(t => t.count > 0);
                    setSelectedType(firstTypeWithItems ? firstTypeWithItems.type : availableTypes[0].type);
                }
            }
        } else {
            setSelectedType(null);
            initializedRef.current = false;
        }
    }, [availableTypes, selectedType]);

    // Now handle conditional rendering after all hooks
    if (!installId) {
        return (
            <WithAdminBodyLayout Icon={BookOpen} name="Spec Documentation" description="View space specification">
                <div className="flex items-center justify-center h-64">
                    <div className="text-center">
                        <BookOpen className="w-16 h-16 text-gray-400 mx-auto mb-4" />
                        <p className="text-gray-500">Install ID not provided</p>
                    </div>
                </div>
            </WithAdminBodyLayout>
        );
    }

    if (loader.loading) {
        return (
            <WithAdminBodyLayout Icon={BookOpen} name="Spec Documentation" description="View space specification">
                <div className="flex items-center justify-center h-64">
                    <div className="text-center">
                        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto mb-4"></div>
                        <p className="text-gray-500">Loading specification...</p>
                    </div>
                </div>
            </WithAdminBodyLayout>
        );
    }

    if (loader.error) {
        return (
            <WithAdminBodyLayout Icon={BookOpen} name="Spec Documentation" description="View space specification">
                <div className="flex items-center justify-center h-64">
                    <div className="text-center">
                        <p className="text-red-500 mb-4">Error loading specification</p>
                        <p className="text-gray-500 text-sm">{String(loader.error)}</p>
                    </div>
                </div>
            </WithAdminBodyLayout>
        );
    }

    if (!spec) {
        return (
            <WithAdminBodyLayout Icon={BookOpen} name="Spec Documentation" description="View space specification">
                <div className="flex items-center justify-center h-64">
                    <div className="text-center">
                        <p className="text-gray-500">No specification data available</p>
                    </div>
                </div>
            </WithAdminBodyLayout>
        );
    }

    const renderEmptyState = (icon: any, title: string) => {
        const Icon = icon;
        return (
            <SpecSection title={title} icon={Icon}>
                <div className="text-center py-12 bg-white rounded-lg border border-gray-200">
                    <Icon className="w-16 h-16 text-gray-300 mx-auto mb-4" />
                    <p className="text-gray-500">No {title.toLowerCase()} defined in this specification</p>
                </div>
            </SpecSection>
        );
    };

    const renderContent = () => {
        if (!selectedType) {
            return (
                <div className="text-center py-12">
                    <BookOpen className="w-16 h-16 text-gray-300 mx-auto mb-4" />
                    <p className="text-gray-500">No specification data available</p>
                </div>
            );
        }

        switch (selectedType) {
            case 'models':
                if (!spec?.models || spec.models.length === 0) {
                    return renderEmptyState(Layers, 'Models');
                }
                return (
                    <SpecSection title="Models" icon={Layers}>
                        <div className="grid grid-cols-1 gap-4">
                            {spec.models.map((model, idx) => (
                                <div key={idx} className="bg-white rounded-lg border border-gray-200 p-5">
                                    <h3 className="text-lg font-semibold text-gray-900 mb-2">{model.name}</h3>
                                    {model.description && (
                                        <p className="text-sm text-gray-600 mb-3">{model.description}</p>
                                    )}
                                    {model.schema && (
                                        <SchemaViewer schema={model.schema} models={spec.models} />
                                    )}
                                </div>
                            ))}
                        </div>
                    </SpecSection>
                );

            case 'scopes':
                if (!spaceSpec?.scopes || spaceSpec.scopes.length === 0) {
                    return renderEmptyState(Box, 'Scopes');
                }
                return (
                    <SpecSection title="Scopes" icon={Box}>
                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                            {spaceSpec.scopes.map((scope, idx) => (
                                <div key={idx} className="bg-white rounded-lg border border-gray-200 p-5">
                                    <h3 className="text-lg font-semibold text-gray-900 mb-2">{scope.name}</h3>
                                    {scope.description && (
                                        <p className="text-sm text-gray-600">{scope.description}</p>
                                    )}
                                </div>
                            ))}
                        </div>
                    </SpecSection>
                );

            case 'events_outputs':
                if (!spaceSpec?.events_outputs || spaceSpec.events_outputs.length === 0) {
                    return renderEmptyState(Zap, 'Event Outputs');
                }
                return (
                    <SpecSection title="Event Outputs" icon={Zap}>
                        <div className="grid grid-cols-1 gap-4">
                            {spaceSpec.events_outputs.map((event, idx) => (
                                <SpecCard
                                    key={idx}
                                    title={event.name}
                                    description={event.description}
                                    schema={event.schema}
                                    models={spec?.models || []}
                                    schemaFile={event.schema_file}
                                />
                            ))}
                        </div>
                    </SpecSection>
                );

            case 'event_slots':
                if (!spaceSpec?.event_slots || spaceSpec.event_slots.length === 0) {
                    return renderEmptyState(Code, 'Event Slots');
                }
                return (
                    <SpecSection title="Event Slots" icon={Code}>
                        <div className="grid grid-cols-1 gap-4">
                            {spaceSpec.event_slots.map((slot, idx) => (
                                <SpecCard
                                    key={idx}
                                    title={slot.name}
                                    description={slot.description}
                                    schema={slot.schema}
                                    models={spec?.models || []}
                                    schemaFile={slot.schema_file}
                                />
                            ))}
                        </div>
                    </SpecSection>
                );

            case 'apis':
                if (!spaceSpec?.APIs || spaceSpec.APIs.length === 0) {
                    return renderEmptyState(Code, 'APIs');
                }
                return (
                    <SpecSection title="APIs" icon={Code}>
                        <div className="grid grid-cols-1 gap-4">
                            {spaceSpec.APIs.map((api, idx) => (
                                <SpecCard
                                    key={idx}
                                    title={api.name}
                                    description={api.description}
                                    schema={api.schema}
                                    models={spec?.models || []}
                                    schemaFile={api.schema_file}
                                />
                            ))}
                        </div>
                    </SpecSection>
                );

            case 'blocks':
                if (!spaceSpec?.blocks || spaceSpec.blocks.length === 0) {
                    return renderEmptyState(Box, 'Blocks');
                }
                return (
                    <SpecSection title="Blocks" icon={Box}>
                        <div className="grid grid-cols-1 gap-4">
                            {spaceSpec.blocks.map((block, idx) => (
                                <SpecCard
                                    key={idx}
                                    title={block.name}
                                    description={block.description}
                                    schema={block.schema}
                                    models={spec?.models || []}
                                    schemaFile={block.schema_file}
                                />
                            ))}
                        </div>
                    </SpecSection>
                );

            default:
                return null;
        }
    };

    return (
        <WithAdminBodyLayout Icon={BookOpen} name="Spec Documentation" description="View space specification">
            <div className="flex flex-1 w-full min-h-0">
                {/* Sidebar */}
                <div className="w-64 bg-white border-r border-gray-200 flex-shrink-0 flex flex-col">
                    <div className="p-4 border-b border-gray-200">
                        <h3 className="text-sm font-semibold text-gray-500 uppercase tracking-wider">Spec Types</h3>
                    </div>
                    <nav className="p-2 flex-1 overflow-y-auto">
                        {availableTypes.map((item) => {
                            const Icon = item.icon;
                            const isActive = selectedType === item.type;
                            const isEmpty = item.count === 0;
                            return (
                                <button
                                    key={item.type}
                                    onClick={() => setSelectedType(item.type)}
                                    className={`w-full flex items-center justify-between px-3 py-2.5 rounded-lg mb-1 transition-colors ${
                                        isActive
                                            ? 'bg-blue-50 text-blue-700 border border-blue-200'
                                            : isEmpty
                                            ? 'text-gray-400 hover:bg-gray-50 hover:text-gray-600'
                                            : 'text-gray-700 hover:bg-gray-50'
                                    }`}
                                >
                                    <div className="flex items-center gap-3">
                                        <Icon className={`w-4 h-4 ${
                                            isActive 
                                                ? 'text-blue-600' 
                                                : isEmpty 
                                                ? 'text-gray-300' 
                                                : 'text-gray-500'
                                        }`} />
                                        <span className={`text-sm font-medium ${isEmpty ? 'opacity-60' : ''}`}>
                                            {item.label}
                                        </span>
                                    </div>
                                    <span className={`text-xs px-2 py-0.5 rounded-full ${
                                        isActive 
                                            ? 'bg-blue-100 text-blue-700' 
                                            : isEmpty 
                                            ? 'bg-gray-50 text-gray-400' 
                                            : 'bg-gray-100 text-gray-600'
                                    }`}>
                                        {item.count}
                                    </span>
                                </button>
                            );
                        })}
                    </nav>
                    {spaceSpecKeys.length > 0 && (
                        <div className="p-4 border-t border-gray-200">
                            <div className="bg-blue-50 border border-blue-200 rounded-lg p-3">
                                <div className="flex items-center gap-2 mb-1">
                                    <Box className="w-4 h-4 text-blue-600" />
                                    <span className="text-xs font-semibold text-blue-900">Space</span>
                                </div>
                                <p className="text-xs text-blue-700 font-medium">{firstSpaceKey}</p>
                            </div>
                        </div>
                    )}
                </div>

                {/* Main Content */}
                <div className="flex-1 overflow-y-auto min-w-0">
                    <div className="w-full max-w-none px-8 py-8">
                        {renderContent()}
                    </div>
                </div>
            </div>
        </WithAdminBodyLayout>
    );
}


