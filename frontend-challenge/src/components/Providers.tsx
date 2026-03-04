'use client';

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { useCartStore } from '@/store/useCartStore';

function CartHydration() {
    const hydrated = React.useRef(false);
    React.useEffect(() => {
        if (!hydrated.current) {
            useCartStore.persist.rehydrate();
            hydrated.current = true;
        }
    }, []);
    return null;
}

export default function Providers({ children }: { children: React.ReactNode }) {
    const [queryClient] = React.useState(() => new QueryClient({
        defaultOptions: {
            queries: {
                staleTime: 60 * 1000,
                retry: 2,
            },
        },
    }));

    return (
        <QueryClientProvider client={queryClient}>
            <CartHydration />
            {children}
        </QueryClientProvider>
    );
}
