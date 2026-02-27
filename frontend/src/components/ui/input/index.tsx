'use client';
import type { InputHTMLAttributes } from 'react';
import clsx from 'clsx';

type Props = InputHTMLAttributes<HTMLInputElement>;

export default function Input({ className, ...rest }: Props) {
  return (
    <input
      className={clsx(
        'px-4 py-2',
        'border border-gray-300',
        'rounded-full',
        'focus:outline-none',
        'focus:ring-2 focus:ring-blue-500',
        className
      )}
      {...rest}
    />
  );
}