'use client';
import type { ButtonHTMLAttributes } from 'react';
import clsx from 'clsx';

type Props = ButtonHTMLAttributes<HTMLButtonElement>;

export default function Button({ className, children, ...rest }: Props) {
  return (
    <button
      className={clsx(
        'px-8 py-2 leading-6',
        'bg-blue-500 text-white',
        'rounded-full',
        'cursor-pointer',
        'font-medium tracking-[0.02em]',
        'inline-flex items-center justify-center',
        'relative shadow',
        //Hover effect
        'transition',
        'hover:bg-blue-600 hover:shadow-md',
        //Focus
        'outline-none',
        'ring-blue-500/70 ring-offset-2',
        'focus-visible:ring-2 focus:scale-[0.98]',
        //Disabled
        'disabled:bg-blue-500/50 disabled:cursor-not-allowed',
        className
      )}
      {...rest}
    >
      {children}
    </button>
  );
}
