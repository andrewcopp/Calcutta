import React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';

import { cn } from '../../lib/cn';

const cardVariants = cva('bg-card rounded-lg', {
  variants: {
    variant: {
      default: 'border border-border shadow-sm',
      elevated: 'border border-border shadow-md',
      accent: 'border border-border border-l-4 border-l-primary shadow-sm',
    },
    padding: {
      default: 'p-6',
      compact: 'p-4',
      none: '',
    },
  },
  defaultVariants: {
    variant: 'default',
    padding: 'default',
  },
});

type CardProps = React.HTMLAttributes<HTMLDivElement> & VariantProps<typeof cardVariants>;

export function Card({ className, variant, padding, ...props }: CardProps) {
  return <div className={cn(cardVariants({ variant, padding }), className)} {...props} />;
}
