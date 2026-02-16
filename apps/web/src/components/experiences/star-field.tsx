"use client";

import { useState } from "react";
import type { UseFormRegisterReturn } from "react-hook-form";

interface StarFieldProps {
  label: string;
  icon: string;
  placeholder: string;
  guide: string;
  maxLength: number;
  defaultValue?: string;
  registration: UseFormRegisterReturn;
  error?: string;
}

export function StarField({
  label,
  icon,
  placeholder,
  guide,
  maxLength,
  defaultValue = "",
  registration,
  error,
}: StarFieldProps) {
  const [length, setLength] = useState(defaultValue.length);

  const { onChange: formOnChange, name, ...restRegistration } = registration;
  const fieldId = `star-${name}`;

  return (
    <div className="space-y-1.5">
      <label htmlFor={fieldId} className="flex items-center gap-1.5 text-sm font-medium text-foreground/80">
        <span>{icon}</span>
        <span>{label}</span>
      </label>
      <p id={`${fieldId}-guide`} className="text-xs text-muted-foreground/60">{guide}</p>
      <textarea
        id={fieldId}
        name={name}
        {...restRegistration}
        onChange={(e) => {
          setLength(e.target.value.length);
          formOnChange(e);
        }}
        placeholder={placeholder}
        rows={4}
        maxLength={maxLength}
        aria-describedby={`${fieldId}-guide`}
        className="w-full rounded-lg border border-border bg-input px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground/50 focus:outline-none focus:ring-2 focus:ring-ring resize-none"
      />
      <div className="flex items-center justify-between">
        {error && <p role="alert" className="text-xs text-red-500">{error}</p>}
        <p
          className={`ml-auto text-xs ${
            length > maxLength ? "text-red-500 font-medium" : "text-muted-foreground/60"
          }`}
          aria-live="polite"
          data-testid="char-counter"
        >
          {length}/{maxLength}
        </p>
      </div>
    </div>
  );
}
