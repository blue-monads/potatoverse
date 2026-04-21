"use client";
import React, { useEffect, useState, useCallback, useRef } from 'react';
import { Database, Table2, ChevronDown, Columns3, RefreshCw, ChevronsUp } from 'lucide-react';
import { useSearchParams, useRouter } from 'next/navigation';
import WithAdminBodyLayout from '@/contain/Layouts/WithAdminBodyLayout';
import TabView from './TabView';
import TablesViewContent from './TablesViewContent';
import LiveQueryViewContent from './LiveQueryViewContent';
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
    const [activeView, setActiveView] = useState<'tables' | 'livequery'>('tables');

    const tablesLoader = useSimpleDataLoader<SpaceDataTable[]>({
        loader: () => listSpaceDataTables(installId),
        ready: true,
    });

    const [columns, setColumns] = useState<SpaceDataColumn[]>([]);
    const [rows, setRows] = useState<Record<string, any>[]>([]);
    const [dataLoading, setDataLoading] = useState(false);
    const [dataError, setDataError] = useState<string | null>(null);
    const [pickerOpen, setPickerOpen] = useState(false);
    const [reachedEnd, setReachedEnd] = useState(false);
    const [currentOffset, setCurrentOffset] = useState(0);
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
        router.push(`?${params.toString()}`);
        setPickerOpen(false);
        setRows([]);
        setColumns([]);
        setCurrentOffset(0);
        setReachedEnd(false);
    };

    const loadInitial = useCallback(async () => {
        if (!selectedTable) return;
        setDataLoading(true);
        setDataError(null);
        setCurrentOffset(0);
        setReachedEnd(false);
        try {
            const [colsRes, rowsRes] = await Promise.all([
                getSpaceDataTableColumns(installId, selectedTable),
                querySpaceDataTable(installId, selectedTable, 0, PAGE_LIMIT),
            ]);
            setColumns(colsRes.data || []);
            const newRows = rowsRes.data || [];
            setRows(newRows);
            setCurrentOffset(newRows.length);
            if (newRows.length < PAGE_LIMIT) setReachedEnd(true);
        } catch (err: any) {
            setDataError(err.message || 'Failed to load table data');
            setColumns([]);
            setRows([]);
        } finally {
            setDataLoading(false);
        }
    }, [installId, selectedTable]);

    const loadMore = useCallback(async () => {
        if (!selectedTable || dataLoading) return;
        setDataLoading(true);
        try {
            const rowsRes = await querySpaceDataTable(installId, selectedTable, currentOffset, PAGE_LIMIT);
            const newRows = rowsRes.data || [];
            if (newRows.length < PAGE_LIMIT) setReachedEnd(true);
            setRows(prev => [...prev, ...newRows]);
            setCurrentOffset(prev => prev + newRows.length);
        } catch (err: any) {
            setDataError(err.message || 'Failed to load more data');
        } finally {
            setDataLoading(false);
        }
    }, [installId, selectedTable, currentOffset, dataLoading]);

    const goToStart = () => {
        setRows([]);
        setCurrentOffset(0);
        setReachedEnd(false);
        loadInitialRef.current?.();
    };

    const loadInitialRef = useRef(loadInitial);
    loadInitialRef.current = loadInitial;

    useEffect(() => {
        if (selectedTable) {
            loadInitial();
        } else {
            setColumns([]);
            setRows([]);
            setCurrentOffset(0);
            setReachedEnd(false);
        }
    }, [selectedTable, loadInitial]);

    const columnNames = columns.length > 0
        ? columns.map(c => c.name)
        : rows.length > 0
            ? Object.keys(rows[0])
            : [];

    const tablesViewProps = {
        selectedTable,
        pickerOpen,
        onPickerToggle: () => setPickerOpen(!pickerOpen),
        pickerRef,
        tables,
        tablesLoading: tablesLoader.loading,
        onTableSelect: setSelectedTable,
        columns,
        dataLoading,
        onRefresh: () => { tablesLoader.reload(); if (selectedTable) loadInitial(); },
        columnNames,
        rows,
        dataError,
        reachedEnd,
        onLoadMore: loadMore,
        onGoToStart: goToStart,
        CellValue,
    };

    return (
        <WithAdminBodyLayout
            Icon={Database}
            name="Data Explorer"
            description="Browse tables and data for this space"
            variant="none"
        >
            <TabView 
                activeView={activeView}
                onViewChange={setActiveView}
                TablesViewComponent={TablesViewContent}
                tablesProps={tablesViewProps}
                LiveQueryViewComponent={LiveQueryViewContent}
            />
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
