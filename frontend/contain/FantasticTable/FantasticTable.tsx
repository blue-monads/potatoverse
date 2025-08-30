import { useEffect, useRef, useState } from "react";
import { EllipsisVertical } from "lucide-react";

interface PropsType {
  columns: ColumnDef[];
  data: any[];
  isLoading?: boolean;
  error?: string;
  noDataMessage?: string;
  onRowClick?: (rowData: any) => void;
  captionText?: string;
  actions?: Actions[];
  classNamesContainer?: string;
  classNamesTable?: string;
  classNamesTableCaption?: string;
  classNamesTableHead?: string;
  classNamesTableBody?: string;
  classNamesTableRow?: string;
  classNamesTableCell?: string;

}

interface ColumnDef {
  title: string;
  key: string;
  icon?: React.ReactNode | string;
  render?: (data: any) => React.ReactNode;
  searchable?: boolean;
}

interface Actions {
  label: string;
  onClick: (rowData: any) => void;
  icon?: React.ReactNode | string;
  dropdown?: boolean;
  className?: string;
}

const FantasticTable = (props: PropsType) => {
  const dropDownActions = props.actions
    ? props.actions.filter((action) => action.dropdown)
    : [];
  const notDropDownActions = props.actions
    ? props.actions.filter((action) => !action.dropdown)
    : [];

  return (
    <div className="table-wrap">
      <table className={`table caption-bottom ${props.classNamesTable}`}>
        <caption className={`pt-4 ${props.classNamesTableCaption}`}>{props.captionText}</caption>
        <thead className={`${props.classNamesTableHead}`}>
          <tr>
            {props.columns.map((col) => (
              <th key={col.key}>
                {col.icon && (
                  typeof col.icon === "string" ? (
                    <span
                      className="inline-block mr-1"
                      dangerouslySetInnerHTML={{ __html: col.icon }}
                    />
                  ) : (
                    <span className="inline-block mr-1">{col.icon}</span>
                  )
                )}
                {col.title}
              </th>
            ))}

            {props.actions && (
              <>
                <td>Actions</td>
              </>
            )}
          </tr>
        </thead>

        <tbody className={`[&>tr]:hover:preset-tonal-primary ${props.classNamesTableBody}`}>
          {props.isLoading ? (
            <tr className={`${props.classNamesTableRow}`}>
              <td colSpan={props.columns.length} className="text-center py-4">
                Loading...
              </td>
            </tr>
          ) : props.error ? (
            <tr className={`${props.classNamesTableRow}`}>
              <td
                colSpan={props.columns.length}
                className="text-center py-4 text-red-500"
              >
                {props.error}
              </td>
            </tr>
          ) : props.data.length === 0 ? (
            <tr className={`${props.classNamesTableRow}`}>
              <td colSpan={props.columns.length} className="text-center py-4">
                {props.noDataMessage || "No data available."}
              </td>
            </tr>
          ) : (
            props.data.map((row, rowIndex) => {
              const [isDropdownOpen, setIsDropdownOpen] = useState(false);
              const [buttonRect, setButtonRect] = useState<DOMRect | null>(
                null
              );
              const buttonRef = useRef<HTMLButtonElement>(null);

              const handleToggleDropdown = () => {
                if (!isDropdownOpen && buttonRef.current) {
                  const rect = buttonRef.current.getBoundingClientRect();
                  setButtonRect(rect);
                }
                setIsDropdownOpen(!isDropdownOpen);
              };

              useEffect(() => {
                const handleClickOutside = (event: MouseEvent) => {
                  if (
                    isDropdownOpen &&
                    buttonRef.current &&
                    !buttonRef.current.contains(event.target as Node)
                  ) {
                    setIsDropdownOpen(false);
                  }
                };

                const handleScroll = () => {
                  if (isDropdownOpen && buttonRef.current) {
                    const rect = buttonRef.current.getBoundingClientRect();
                    setButtonRect(rect);
                  }
                };

                document.addEventListener("mousedown", handleClickOutside);
                window.addEventListener("scroll", handleScroll, true);
                window.addEventListener("resize", handleScroll);

                return () => {
                  document.removeEventListener("mousedown", handleClickOutside);
                  window.removeEventListener("scroll", handleScroll, true);
                  window.removeEventListener("resize", handleScroll);
                };
              }, [isDropdownOpen]);

              return (
                <tr
                  key={rowIndex}
                  onClick={() => props.onRowClick && props.onRowClick(row)}
                  className={props.onRowClick ? "cursor-pointer" : ""}
                >
                  {props.columns.map((col) => (
                    <td key={col.key} className={`${props.classNamesTableCell}`}>
                      {col.render ? col.render(row[col.key]) : row[col.key]}
                    </td>
                  ))}

                  {props.actions && (
                    <td className={`text-right ${props.classNamesTableCell}`}>
                      {notDropDownActions.map((action, actionIndex) => (
                        <button
                          key={actionIndex}
                          onClick={(e) => {
                            e.stopPropagation();
                            action.onClick(row);
                          }}
                          className={`btn btn-sm btn-base preset-filled ${
                            action.className || ""
                          }`}
                        >
                          {action.icon && (
                            <span className="mr-1">{action.icon}</span>
                          )}
                          {action.label}
                        </button>
                      ))}

                      {dropDownActions.length > 0 && (
                        <div className="relative inline-block text-left">
                          <div>
                            <button
                              type="button"
                              className="inline-flex"
                              id="menu-button"
                              aria-expanded="true"
                              aria-haspopup="true"
                              onClick={(e) => {
                                e.stopPropagation();
                                handleToggleDropdown();
                              }}
                              ref={buttonRef}
                            >
                              <EllipsisVertical className="h-5 w-5" />
                            </button>
                          </div>

                          {isDropdownOpen && buttonRect && (
                            <div
                              className="origin-top-right fixed right-0 mt-2 w-32  bg-white shadow-lg "
                              style={{
                                top: buttonRect.bottom + 4,
                                left: buttonRect.right - 130,
                              }}
                              role="menu"
                              aria-orientation="vertical"
                              aria-labelledby="menu-button"
                              tabIndex={-1}
                            >
                              {dropDownActions.map((action, actionIndex) => (
                                <button
                                  key={actionIndex}
                                  onClick={(e) => {
                                    e.stopPropagation();
                                    action.onClick(row);
                                    setIsDropdownOpen(false);
                                  }}
                                  className={`w-full text-left px-3 py-2 text-sm first:rounded-t-lg last:rounded-b-lg text-gray-700 hover:text-blue-600 transition-colors hover:bg-gray-200 cursor-pointer ${
                                    action.className || ""
                                  }`}
                                  role="menuitem"
                                  tabIndex={-1}
                                  id={`menu-item-${actionIndex}`}
                                >
                                  <div className="inline-flex items-center gap-2">
                                    {typeof action.icon === "string" ? (
                                      <span
                                        dangerouslySetInnerHTML={{
                                          __html: action.icon,
                                        }}
                                      />
                                    ) : (
                                      action.icon
                                    )}
                                    {action.label}
                                  </div>
                                </button>
                              ))}
                            </div>
                          )}
                        </div>
                      )}
                    </td>
                  )}
                </tr>
              );
            })
          )}
        </tbody>
      </table>
    </div>
  );
};

export default FantasticTable;
