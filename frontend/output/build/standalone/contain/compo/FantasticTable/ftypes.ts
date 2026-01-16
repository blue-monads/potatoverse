
export interface Actions {
    label: string;
    onClick: (rowData: any) => void;
    icon?: React.ReactNode | string;
    dropdown?: boolean;
    className?: string;
    color?: "primary" | "secondary" | "success" | "danger" | "warning"
  }