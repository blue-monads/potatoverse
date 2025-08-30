import React from "react";
import DropdownActions from "./DropdownActions";
import { Actions } from "./ftypes";

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
            props.data.map((row, rowIndex) => (
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
                        className={`btn btn-sm  preset-filled ${
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
                      <DropdownActions
                        actions={dropDownActions}
                        rowData={row}
                        classNamesTableCell={props.classNamesTableCell}
                      />
                    )}
                  </td>
                )}
              </tr>
            ))
          )}
        </tbody>
      </table>
    </div>
  );
};

export default FantasticTable;
