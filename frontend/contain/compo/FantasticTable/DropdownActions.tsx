import { useEffect, useRef, useState } from "react";
import { EllipsisVertical } from "lucide-react";
import { Actions } from "./ftypes";


interface DropdownActionsProps {
  actions: Actions[];
  rowData: any;
  classNamesTableCell?: string;
}

const DropdownActions = ({ actions, rowData, classNamesTableCell }: DropdownActionsProps) => {
  const [isDropdownOpen, setIsDropdownOpen] = useState(false);
  const [buttonRect, setButtonRect] = useState<DOMRect | null>(null);
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
          className="origin-top-right fixed right-0 mt-2 w-32 bg-white shadow-lg"
          style={{
            top: buttonRect.bottom + 4,
            left: buttonRect.right - 130,
          }}
          role="menu"
          aria-orientation="vertical"
          aria-labelledby="menu-button"
          tabIndex={-1}
        >
          {actions.map((action, actionIndex) => (
            <button
              key={actionIndex}
              onClick={(e) => {
                e.stopPropagation();
                action.onClick(rowData);
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
  );
};

export default DropdownActions;
