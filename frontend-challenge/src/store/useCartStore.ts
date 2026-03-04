import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { Product, CartItem } from '../types';

export const MAX_QUANTITY = 99;

interface CartState {
    items: CartItem[];
    addItem: (product: Product) => void;
    removeItem: (productId: string) => void;
    updateQuantity: (productId: string, quantity: number) => void;
    clearCart: () => void;
    syncPrices: (currentProducts: Product[]) => string[];
}

export const useCartStore = create<CartState>()(
    persist(
        (set, get) => ({
            items: [],

            addItem: (product: Product) => {
                set((state) => {
                    const existingItem = state.items.find((item) => item.productId === product.id);
                    if (existingItem) {
                        return {
                            items: state.items.map((item) =>
                                item.productId === product.id
                                    ? { ...item, quantity: Math.min(item.quantity + 1, MAX_QUANTITY) }
                                    : item
                            ),
                        };
                    }
                    return {
                        items: [...state.items, { productId: product.id, quantity: 1, product }],
                    };
                });
            },

            removeItem: (productId: string) => {
                set((state) => ({
                    items: state.items.filter((item) => item.productId !== productId),
                }));
            },

            updateQuantity: (productId: string, quantity: number) => {
                if (quantity <= 0) {
                    get().removeItem(productId);
                    return;
                }
                const clamped = Math.min(quantity, MAX_QUANTITY);
                set((state) => ({
                    items: state.items.map((item) =>
                        item.productId === productId ? { ...item, quantity: clamped } : item
                    ),
                }));
            },

            clearCart: () => set({ items: [] }),

            syncPrices: (currentProducts: Product[]) => {
                const warnings: string[] = [];
                const currentState = get();

                let hasChanges = false;
                const newItems = currentState.items.map((item) => {
                    const currentProduct = currentProducts.find(p => p.id === item.productId);

                    if (!currentProduct) {
                        hasChanges = true;
                        warnings.push(`${item.product.name} is no longer available and was removed from your cart.`);
                        return null;
                    }

                    if (currentProduct.price > item.product.price) {
                        hasChanges = true;
                        warnings.push(`The price of ${currentProduct.name} has increased to $${currentProduct.price.toFixed(2)}.`);
                        return { ...item, product: currentProduct };
                    }

                    if (currentProduct.price !== item.product.price || currentProduct.name !== item.product.name) {
                        hasChanges = true;
                        return { ...item, product: currentProduct };
                    }

                    return item;
                }).filter(Boolean) as CartItem[];

                if (hasChanges) {
                    set({ items: newItems });
                }

                return warnings;
            },
        }),
        {
            name: 'shopping-cart-storage',
            skipHydration: true,
        }
    )
);

// Selectors — use these instead of subscribing to the entire store
export const selectCartItem = (productId: string) =>
    (state: CartState) => state.items.find(i => i.productId === productId);

export const selectTotalItems = (state: CartState) =>
    state.items.reduce((sum, item) => sum + item.quantity, 0);

export const selectSubtotal = (state: CartState) =>
    state.items.reduce((sum, item) => sum + item.product.price * item.quantity, 0);
