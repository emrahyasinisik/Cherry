"use client";

import { useEffect, useRef, type ChangeEvent, type ClipboardEvent, type KeyboardEvent } from "react";

type CodeInputsProps = {
  value: string;
  onChange: (next: string) => void;
  disabled?: boolean;
  error?: boolean;
};

export function CodeInputs({ value, onChange, disabled, error }: CodeInputsProps) {
  const digits = toCells(value);
  const refs = useRef<Array<HTMLInputElement | null>>([]);

  useEffect(() => {
    refs.current[0]?.focus();
  }, []);

  function setAt(index: number, char: string) {
    const next = [...digits];
    next[index] = char;
    onChange(next.join(""));
  }

  function handleChange(index: number, event: ChangeEvent<HTMLInputElement>) {
    const raw = event.target.value.replace(/\D/g, "");
    if (!raw) {
      setAt(index, "");
      return;
    }
    if (raw.length > 1) {
      const merged = (digits.join("").slice(0, index) + raw).replace(/\D/g, "").slice(0, 6);
      onChange(merged);
      const focusAt = Math.min(index + raw.length, 5);
      refs.current[focusAt]?.focus();
      return;
    }
    setAt(index, raw);
    if (index < 5) {
      refs.current[index + 1]?.focus();
    }
  }

  function handleKeyDown(index: number, event: KeyboardEvent<HTMLInputElement>) {
    if (event.key === "Backspace" && !digits[index] && index > 0) {
      refs.current[index - 1]?.focus();
    }
  }

  function handlePaste(event: ClipboardEvent<HTMLInputElement>) {
    event.preventDefault();
    const pasted = event.clipboardData.getData("text").replace(/\D/g, "").slice(0, 6);
    onChange(pasted);
    refs.current[Math.min(pasted.length, 5)]?.focus();
  }

  return (
    <div className="flex justify-center gap-2">
      {digits.map((digit, index) => (
        <input
          key={index}
          ref={(node) => {
            refs.current[index] = node;
          }}
          type="text"
          inputMode="numeric"
          autoComplete={index === 0 ? "one-time-code" : "off"}
          maxLength={1}
          name={index === 0 ? "otp" : undefined}
          aria-label={`Kod ${index + 1}`}
          disabled={disabled}
          value={digit}
          onChange={(event) => handleChange(index, event)}
          onKeyDown={(event) => handleKeyDown(index, event)}
          onPaste={handlePaste}
          className="cherry-focus cherry-code-cell h-12 w-10 rounded-md border border-input bg-transparent text-center font-mono text-lg outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50"
          data-filled={digit.length > 0}
          data-error={error}
        />
      ))}
    </div>
  );
}

function toCells(value: string): string[] {
  const cleaned = value.replace(/\D/g, "").slice(0, 6);
  return Array.from({ length: 6 }, (_, index) => cleaned[index] ?? "");
}
