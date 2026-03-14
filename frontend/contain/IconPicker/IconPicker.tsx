import { useState, useEffect, useMemo, useCallback, useRef } from "react";


const ICONS_PER_PAGE = 120;

type PropTypes = {
  onSelect: (icon: string) => void;
}


export const IconPicker = (props: PropTypes) => {
  const { onSelect } = props;
  
  const [icons, setIcons] = useState<string[]>([]);
  const [search, setSearch] = useState("");
  const [selected, setSelected] = useState<string | null>(null);
  const [page, setPage] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    fetch("/zz/static/fontawesome/icon_map.json")
      .then((r) => r.json())
      .then(setIcons)
      .catch(console.error);
  }, []);

  const filtered = useMemo(
    () =>
      search.trim()
        ? icons.filter((name) => name.includes(search.trim().toLowerCase()))
        : icons,
    [icons, search]
  );

  const totalPages = Math.ceil(filtered.length / ICONS_PER_PAGE);
  const pageIcons = filtered.slice(
    page * ICONS_PER_PAGE,
    (page + 1) * ICONS_PER_PAGE
  );

  useEffect(() => {
    setPage(0);
  }, [search]);

  const handleSelect = useCallback((name: string) => {
    setSelected((prev) => (prev === name ? null : name));
    onSelect(name);
  }, [onSelect]);

  return (
    <div className="min-h-screen bg-surface-50 p-6">
      <div className="max-w-5xl mx-auto">
        <h1 className="text-2xl font-bold mb-1">Icon Picker</h1>
        <p className="text-sm text-gray-500 mb-4">
          {icons.length} Font Awesome icons &middot; click to select
        </p>

        {/* search + selected preview */}
        <div className="flex items-center gap-4 mb-5">
          <div className="relative flex-1">
            <i className="fa-solid fa-magnifying-glass absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 text-sm" />
            <input
              ref={inputRef}
              type="text"
              placeholder="Search icons…"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full pl-9 pr-4 py-2 rounded-lg border border-gray-300 bg-white text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
          </div>

          {selected && (
            <div className="flex items-center gap-2 px-4 py-2 rounded-lg bg-blue-50 border border-blue-200 text-sm shrink-0">
              <i className={`fa-solid fa-${selected} text-blue-600 text-lg`} />
              <code className="text-blue-800 font-mono text-xs">
                fa-{selected}
              </code>
              <button
                onClick={() => setSelected(null)}
                className="ml-1 text-blue-400 hover:text-blue-600"
              >
                <i className="fa-solid fa-xmark text-xs" />
              </button>
            </div>
          )}
        </div>

        <p className="text-xs text-gray-400 mb-3">
          Showing {pageIcons.length} of {filtered.length} results
          {totalPages > 1 && ` · page ${page + 1}/${totalPages}`}
        </p>

        {/* icon grid */}
        <div className="grid grid-cols-6 sm:grid-cols-8 md:grid-cols-10 lg:grid-cols-12 gap-2">
          {pageIcons.map((name) => (
            <button
              key={name}
              onClick={() => handleSelect(name)}
              title={name}
              className={`group flex flex-col items-center justify-center gap-1.5 p-3 rounded-lg border transition-all cursor-pointer
                ${
                  selected === name
                    ? "border-blue-500 bg-blue-50 ring-2 ring-blue-300"
                    : "border-gray-200 bg-white hover:border-gray-300 hover:shadow-sm"
                }`}
            >
              <i
                className={`fa-solid fa-${name} text-xl ${
                  selected === name
                    ? "text-blue-600"
                    : "text-gray-600 group-hover:text-gray-900"
                }`}
              />
              <span className="text-[10px] text-gray-400 truncate w-full text-center leading-tight">
                {name}
              </span>
            </button>
          ))}
        </div>

        {filtered.length === 0 && (
          <div className="text-center py-16 text-gray-400">
            <i className="fa-solid fa-face-sad-tear text-4xl mb-3 block" />
            No icons match &ldquo;{search}&rdquo;
          </div>
        )}

        {/* pagination */}
        {totalPages > 1 && (
          <div className="flex items-center justify-center gap-2 mt-6">
            <button
              onClick={() => setPage((p) => Math.max(0, p - 1))}
              disabled={page === 0}
              className="px-3 py-1.5 text-sm rounded-md border border-gray-300 bg-white hover:bg-gray-50 disabled:opacity-40 disabled:cursor-not-allowed"
            >
              <i className="fa-solid fa-chevron-left mr-1" /> Prev
            </button>
            <span className="text-sm text-gray-500 px-2">
              {page + 1} / {totalPages}
            </span>
            <button
              onClick={() => setPage((p) => Math.min(totalPages - 1, p + 1))}
              disabled={page >= totalPages - 1}
              className="px-3 py-1.5 text-sm rounded-md border border-gray-300 bg-white hover:bg-gray-50 disabled:opacity-40 disabled:cursor-not-allowed"
            >
              Next <i className="fa-solid fa-chevron-right ml-1" />
            </button>
          </div>
        )}
      </div>
    </div>
  );
}