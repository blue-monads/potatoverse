import React from 'react';
import { CapabilityOptionField } from '@/lib';

interface CapabilityOptionFieldInputProps {
    field: CapabilityOptionField;
    value: any;
    onChange: (value: any) => void;
}

export const CapabilityOptionFieldInput = ({
    field,
    value,
    onChange
}: CapabilityOptionFieldInputProps) => {
    const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
        let newValue: any = e.target.value;
        
        if (field.type === 'number') {
            newValue = parseFloat(newValue) || 0;
        } else if (field.type === 'boolean') {
            newValue = (e.target as HTMLInputElement).checked;
        }

        onChange(newValue);
    };

    return (
        <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
                {field.name}
                {field.required && <span className="text-red-500">*</span>}
            </label>
            {field.description && (
                <p className="text-xs text-gray-500 mb-1">{field.description}</p>
            )}
            {field.type === 'textarea' ? (
                <textarea
                    value={value || ''}
                    onChange={handleChange}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                    required={field.required}
                    rows={3}
                />
            ) : field.type === 'select' ? (
                <select
                    value={value || field.default || ''}
                    onChange={handleChange}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                    required={field.required}
                >
                    {field.options.map(opt => (
                        <option key={opt} value={opt}>{opt}</option>
                    ))}
                </select>
            ) : field.type === 'boolean' ? (
                <label className="flex items-center gap-2">
                    <input
                        type="checkbox"
                        checked={value || false}
                        onChange={handleChange}
                        className="w-4 h-4"
                    />
                    <span className="text-sm text-gray-700">Enabled</span>
                </label>
            ) : (
                <input
                    type={field.type === 'api_key' ? 'password' : field.type === 'number' ? 'number' : 'text'}
                    value={value || ''}
                    onChange={handleChange}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                    required={field.required}
                    placeholder={field.default}
                />
            )}
        </div>
    );
};

