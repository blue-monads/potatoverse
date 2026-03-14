"use client";
import React, { useEffect, useState, useCallback, useRef } from 'react';
import { Database, Table2, ChevronLeft, ChevronRight, ChevronDown, Columns3, RefreshCw } from 'lucide-react';
import { useSearchParams, useRouter } from 'next/navigation';
import WithAdminBodyLayout from '@/contain/Layouts/WithAdminBodyLayout';
import {
    listSpaceDataTables,
    getSpaceDataTableColumns,
    querySpaceDataTable,
    SpaceDataTable,
    SpaceDataColumn,
} from '@/lib';
import useSimpleDataLoader from '@/hooks/useSimpleDataLoader';

const PAGE_LIMIT = 50;

export default function Page() {
    const searchParams = useSearchParams();
    const installId = searchParams.get('install_id');

    if (!installId) {
        return <div>Install ID not provided</div>;
    }

    return <DataExplorerPage installId={parseInt(installId)} />;
}

const DataExplorerPage = ({ installId }: { installId: number }) => {
    const searchParams = useSearchParams();
    const router = useRouter();

    const selectedTable = searchParams.get('table') || '';
    const currentOffset = parseInt(searchParams.get('offset') || '0', 10) || 0;

    const tablesLoader = useSimpleDataLoader<SpaceDataTable[]>({
        loader: () => listSpaceDataTables(installId),
        ready: true,
    });

    const [columns, setColumns] = useState<SpaceDataColumn[]>([]);
    const [rows, setRows] = useState<Record<string, any>[]>([]);
    const [dataLoading, setDataLoading] = useState(false);
    const [dataError, setDataError] = useState<string | null>(null);
    const [pickerOpen, setPickerOpen] = useState(false);
    const pickerRef = useRef<HTMLDivElement>(null);

    const tables = tablesLoader.data || [];

    useEffect(() => {
        const handleClick = (e: MouseEvent) => {
            if (pickerOpen && pickerRef.current && !pickerRef.current.contains(e.target as Node)) {
                setPickerOpen(false);
            }
        };
        document.addEventListener('mousedown', handleClick);
        return () => document.removeEventListener('mousedown', handleClick);
    }, [pickerOpen]);

    const setSelectedTable = (table: string) => {
        const params = new URLSearchParams(searchParams.toString());
        if (table) {
            params.set('table', table);
        } else {
            params.delete('table');
        }
        params.delete('offset');
        router.push(`?${params.toString()}`);
        setPickerOpen(false);
    };

    const loadTableData = useCallback(async () => {
        if (!selectedTable) return;
        setDataLoading(true);
        setDataError(null);
        try {
            const [colsRes, rowsRes] = await Promise.all([
                getSpaceDataTableColumns(installId, selectedTable),
                querySpaceDataTable(installId, selectedTable, currentOffset, PAGE_LIMIT),
            ]);
            setColumns(colsRes.data || []);
            setRows(rowsRes.data || []);
        } catch (err: any) {
            setDataError(err.message || 'Failed to load table data');
            setColumns([]);
            setRows([]);
        } finally {
            setDataLoading(false);
        }
    }, [installId, selectedTable, currentOffset]);

    useEffect(() => {
        if (selectedTable) {
            loadTableData();
        } else {
            setColumns([]);
            setRows([]);
        }
    }, [selectedTable, currentOffset, loadTableData]);

    const hasNext = rows.length === PAGE_LIMIT;
    const hasPrevious = currentOffset > 0;

    const navigate = (newOffset: number) => {
        const params = new URLSearchParams(searchParams.toString());
        if (newOffset === 0) {
            params.delete('offset');
        } else {
            params.set('offset', newOffset.toString());
        }
        router.push(`?${params.toString()}`);
    };

    const columnNames = columns.length > 0
        ? columns.map(c => c.name)
        : rows.length > 0
            ? Object.keys(rows[0])
            : [];

    return (
        <WithAdminBodyLayout
            Icon={Database}
            name="Data Explorer"
            description="Browse tables and data for this space"
            variant="none"
        >
            <div className="flex flex-col flex-1 overflow-hidden" style={{ height: 'calc(100vh - 280px)' }}>
                {/* Tables tab bar */}
                <div className="bg-white border-b border-gray-200 shrink-0">
                    <div className="flex items-center gap-0 px-2 overflow-x-auto no-scrollbar">
                        {/* Table picker dropdown */}
                        <div className="relative shrink-0" ref={pickerRef}>
                            <button
                                onClick={() => setPickerOpen(!pickerOpen)}
                                className="flex items-center gap-1.5 px-3 py-2.5 text-sm text-gray-500 hover:text-gray-700 transition-colors"
                            >
                                <Table2 className="w-3.5 h-3.5" />
                                <ChevronDown className={`w-3.5 h-3.5 transition-transform ${pickerOpen ? 'rotate-180' : ''}`} />
                            </button>

                            {pickerOpen && (
                                <div className="absolute top-full left-0 mt-1 w-56 bg-white border border-gray-200 rounded-lg shadow-lg z-30 py-1 max-h-72 overflow-y-auto">
                                    {tablesLoader.loading ? (
                                        <div className="px-3 py-4 text-sm text-gray-400 text-center">Loading...</div>
                                    ) : tables.length === 0 ? (
                                        <div className="px-3 py-4 text-sm text-gray-400 text-center">No tables</div>
                                    ) : (
                                        tables.map((t) => (
                                            <button
                                                key={t.name}
                                                onClick={() => setSelectedTable(t.name)}
                                                className={`w-full text-left px-3 py-2 text-sm flex items-center gap-2 transition-colors ${selectedTable === t.name
                                                    ? 'bg-blue-50 text-blue-700 font-medium'
                                                    : 'text-gray-700 hover:bg-gray-50'
                                                    }`}
                                            >
                                                <Table2 className="w-3.5 h-3.5 shrink-0" />
                                                <span className="truncate">{t.name}</span>
                                            </button>
                                        ))
                                    )}
                                </div>
                            )}
                        </div>

                        <div className="w-px h-5 bg-gray-200 shrink-0" />

                        {/* Table tabs */}
                        {tables.map((t) => (
                            <button
                                key={t.name}
                                onClick={() => setSelectedTable(t.name)}
                                className={`shrink-0 px-3 py-2.5 text-sm font-medium transition-colors whitespace-nowrap border-b-2 ${selectedTable === t.name
                                    ? 'text-blue-600 border-blue-600'
                                    : 'text-gray-500 border-transparent hover:text-gray-700 hover:bg-gray-50'
                                    }`}
                            >
                                {t.name}
                            </button>
                        ))}

                        {/* Right side: refresh + column count */}
                        <div className="ml-auto shrink-0 flex items-center gap-2 pl-2">
                            {selectedTable && columns.length > 0 && (
                                <span className="text-xs text-gray-400 flex items-center gap-1">
                                    <Columns3 className="w-3 h-3" />
                                    {columns.length}
                                </span>
                            )}
                            <button
                                onClick={() => { tablesLoader.reload(); if (selectedTable) loadTableData(); }}
                                disabled={dataLoading}
                                className="p-1.5 text-gray-400 hover:text-gray-600 rounded transition-colors disabled:opacity-50"
                                title="Refresh"
                            >
                                <RefreshCw className={`w-3.5 h-3.5 ${dataLoading ? 'animate-spin' : ''}`} />
                            </button>
                        </div>
                    </div>
                </div>

                {/* Data viewer */}
                <div className="flex-1 flex flex-col overflow-hidden bg-gray-50">
                    {!selectedTable ? (
                        <div className="flex-1 flex items-center justify-center text-gray-400">
                            <div className="text-center">
                                <Database className="w-12 h-12 mx-auto mb-3 opacity-30" />
                                <p className="text-sm">Select a table to view its data</p>
                            </div>
                        </div>
                    ) : (
                        <>
                            {/* Data table */}
                            <div className="flex-1 overflow-auto">
                                {dataLoading && rows.length === 0 ? (
                                    <div className="flex items-center justify-center h-full text-gray-400 text-sm">
                                        Loading...
                                    </div>
                                ) : dataError ? (
                                    <div className="flex items-center justify-center h-full text-red-500 text-sm">
                                        {dataError}
                                    </div>
                                ) : rows.length === 0 ? (
                                    <div className="flex items-center justify-center h-full text-gray-400 text-sm">
                                        No rows found
                                    </div>
                                ) : (
                                    <table className="min-w-full text-sm">
                                        <thead className="bg-gray-100 sticky top-0 z-10">
                                            <tr>
                                                {columnNames.map((col) => {
                                                    const colMeta = columns.find(c => c.name === col);
                                                    return (
                                                        <th
                                                            key={col}
                                                            className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider whitespace-nowrap border-b border-gray-200"
                                                        >
                                                            <div className="flex items-center gap-1">
                                                                {col}
                                                                {colMeta?.primary_key === 1 && (
                                                                    <span className="text-[9px] bg-yellow-100 text-yellow-700 px-1 rounded font-bold">PK</span>
                                                                )}
                                                            </div>
                                                            {colMeta && (
                                                                <span className="text-[10px] text-gray-400 font-normal normal-case">
                                                                    {colMeta.data_type}
                                                                </span>
                                                            )}
                                                        </th>
                                                    );
                                                })}
                                            </tr>
                                        </thead>
                                        <tbody className="bg-white divide-y divide-gray-100">
                                            {rows.map((row, rowIdx) => (
                                                <tr key={row['id'] ?? rowIdx} className="hover:bg-blue-50/30">
                                                    {columnNames.map((col) => (
                                                        <td
                                                            key={col}
                                                            className="px-4 py-2 whitespace-nowrap text-gray-700 max-w-xs truncate border-b border-gray-50"
                                                            title={String(row[col] ?? '')}
                                                        >
                                                            <CellValue value={row[col]} />
                                                        </td>
                                                    ))}
                                                </tr>
                                            ))}
                                        </tbody>
                                    </table>
                                )}
                            </div>

                            {/* Pagination */}
                            {(hasPrevious || hasNext) && (
                                <div className="px-4 py-2.5 bg-white border-t border-gray-200 flex items-center justify-between shrink-0">
                                    <div className="text-xs text-gray-500">
                                        Showing {currentOffset + 1} - {currentOffset + rows.length}
                                    </div>
                                    <div className="flex items-center gap-2">
                                        <button
                                            onClick={() => navigate(Math.max(0, currentOffset - PAGE_LIMIT))}
                                            disabled={!hasPrevious || dataLoading}
                                            className={`flex items-center gap-1 px-3 py-1.5 text-xs font-medium rounded-md transition-colors ${hasPrevious && !dataLoading
                                                ? 'bg-white text-gray-700 border border-gray-300 hover:bg-gray-50'
                                                : 'bg-gray-100 text-gray-400 cursor-not-allowed'
                                                }`}
                                        >
                                            <ChevronLeft className="w-3.5 h-3.5" />
                                            Prev
                                        </button>
                                        <button
                                            onClick={() => navigate(currentOffset + PAGE_LIMIT)}
                                            disabled={!hasNext || dataLoading}
                                            className={`flex items-center gap-1 px-3 py-1.5 text-xs font-medium rounded-md transition-colors ${hasNext && !dataLoading
                                                ? 'bg-white text-gray-700 border border-gray-300 hover:bg-gray-50'
                                                : 'bg-gray-100 text-gray-400 cursor-not-allowed'
                                                }`}
                                        >
                                            Next
                                            <ChevronRight className="w-3.5 h-3.5" />
                                        </button>
                                    </div>
                                </div>
                            )}
                        </>
                    )}
                </div>
            </div>
        </WithAdminBodyLayout>
    );
};

const CellValue = ({ value }: { value: any }) => {
    if (value === null || value === undefined) {
        return <span className="text-gray-300 italic">null</span>;
    }
    if (typeof value === 'boolean') {
        return <span className={value ? 'text-green-600' : 'text-red-500'}>{String(value)}</span>;
    }
    if (typeof value === 'object') {
        return <span className="font-mono text-xs text-gray-500">{JSON.stringify(value)}</span>;
    }
    return <>{String(value)}</>;
};
