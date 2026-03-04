# Frontend (Next.js) Deep-Dive + Implementation Guide

This document explains the frontend codebase **step-by-step** as if you’re new to React and Next.js. It also describes the **major components**, how data flows through the app, and includes a **TODO checklist** for extending/implementing features.

---

## 1) What this frontend is

It’s a simple food ordering UI with:

- A **product list** (fetched from a backend)
- A **cart** that persists in the browser
- A **promo code** field
- A **place order** action (POST request)
- A **confirmation modal** after ordering
- Error/success/warning snackbars

The UI is built with:

- **Next.js App Router** (`src/app`)
- **React** components (`src/components`)
- **MUI (Material UI)** for styling/layout
- **Zustand** for cart state (client-side store)
- **TanStack React Query** for server state (fetching/mutations)
- **Axios** for HTTP calls

---

## 2) Folder structure (what goes where)

Inside `src/`:

- `app/`
  - Next.js App Router entry files (route layout, page, global CSS)
- `components/`
  - Reusable UI building blocks
  - `cart/` cart sidebar + order confirmation modal
  - `products/` product card
  - `Providers.tsx` React Query provider
  - `ThemeRegistry.tsx` MUI theme + Next.js/MUI integration
- `services/`
  - HTTP API functions (`getProducts`, `placeOrder`)
- `store/`
  - Zustand store for cart (`useCartStore.ts`)
- `theme/`
  - MUI theme builder (`getTheme`) used by `ThemeRegistry`
- `types/`
  - Shared TypeScript types (`Product`, `OrderResponse`, etc.)

Key idea:

- **Server state** (data coming from backend) is handled by **React Query**.
- **Client state** (cart items, quantities) is handled by **Zustand**.

---

## 3) Next.js basics used here (beginner-friendly)

### 3.1 App Router

This project uses the App Router, which means:

- `src/app/layout.tsx` defines the global HTML skeleton (wrapping the entire app)
- `src/app/page.tsx` is the “home route” (i.e. `/`)

### 3.2 Server vs Client Components

In App Router, files are **Server Components by default**.

When you see this line:

```ts
"use client";
```

it means:

- This component runs on the browser
- It can use React hooks like `useState`, `useEffect`
- It can access browser-only APIs like `localStorage`

In this app, the main interactive pieces (`page.tsx`, `Cart`, `ProductCard`, `OrderModal`, providers) are client components.

---

## 4) Global app bootstrapping (how the app starts)

### 4.1 `src/app/layout.tsx`

This is the top-level wrapper.

Main responsibilities:

- Loads a Google font (`Red_Hat_Text`) using Next’s font optimization
- Imports `globals.css`
- Wraps all pages with:
  - `ThemeRegistry` (MUI theme + CSS baseline)
  - `Providers` (React Query client)

The important part:

- `ThemeRegistry` ensures MUI styling works correctly in Next App Router.
- `Providers` creates a single `QueryClient` for the app.

### 4.2 `src/components/Providers.tsx` (React Query)

React Query needs a `QueryClientProvider` at the top of your app.

This file:

- Creates a `QueryClient` using `useState(() => new QueryClient(...))`
  - The `useState` trick ensures the client is only created once per browser session
- Sets a default `staleTime` (1 minute)
  - Meaning: fetched data is considered “fresh” for 60 seconds

Concept:

- React Query caches server responses.
- `staleTime` influences when React Query decides it should refetch.

### 4.3 `src/components/ThemeRegistry.tsx` (MUI)

This wraps the app with:

- `AppRouterCacheProvider` (MUI + Next App Router integration)
- `ThemeProvider` (provides `theme` to MUI components)
- `CssBaseline` (normalizes CSS across browsers)

`getTheme(fontFamily)` creates the MUI theme (colors, typography, component styles).

---

## 5) Types (what the backend returns)

File: `src/types/index.ts`

Important types:

- `Product`
  - `id`, `name`, `price`, `category`, optional `image`
- `OrderItem`
  - `productId`, `quantity`
- `OrderRequest`
  - `couponCode?`, `items: OrderItem[]`
- `OrderResponse`
  - `id`, `items`, `products` (expanded product details), totals, discounts
- `CartItem`
  - Extends `OrderItem` and includes `product: Product`

Why TypeScript types matter:

- You get autocomplete and compile-time checks.
- Your UI code knows exactly what fields exist.

---

## 6) API layer (how frontend talks to backend)

File: `src/services/api.ts`

### 6.1 Axios client

```ts
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
const API_KEY = process.env.NEXT_PUBLIC_API_KEY || "apitest";
```

Concepts:

- `NEXT_PUBLIC_*` env vars are exposed to the browser.
- The axios instance is created with:
  - `baseURL`
  - default headers (`Content-Type`, `api_key`)

### 6.2 API functions

- `getProducts()`
  - `GET /api/product`
  - Returns `Product[]`

- `placeOrder(order)`
  - `POST /api/order`
  - Sends `{ couponCode?, items: [{productId, quantity}] }`
  - Returns `OrderResponse`

Pattern:

- Components never call axios directly.
- Components call these functions, which keeps code organized.

---

## 7) Cart state (Zustand) — the “client state”

File: `src/store/useCartStore.ts`

Zustand is a small state management library.

### 7.1 What the store contains

- `items: CartItem[]`
- Actions:
  - `addItem(product)`
  - `removeItem(productId)`
  - `updateQuantity(productId, quantity)`
  - `clearCart()`
- Selectors / computed helpers:
  - `getTotalItems()`
  - `getSubtotal()`
- Sync helper:
  - `syncPrices(currentProducts)`

### 7.2 Persistence

The store uses `persist(...)` middleware:

- Saves cart to `localStorage` with key `shopping-cart-storage`
- So if you refresh the page, cart items remain.

### 7.3 How updates work (immutability)

Inside Zustand actions, you’ll see patterns like:

- `set((state) => ({ items: ... }))`

This returns a **new array** instead of mutating the old one.

Why:

- React rendering relies on detecting changes by reference.

### 7.4 `syncPrices` (important)

`syncPrices` takes the latest product list from the server and:

- Removes items that no longer exist
- Updates product details in cart if changed
- Generates **warnings** if:
  - item removed
  - price increased

This is a defensive feature:

- Your cart is stored locally, but server data can change.

---

## 8) Home page (`src/app/page.tsx`) — end-to-end flow

This is the core “composition” file.

### 8.1 State in this component

Local React state:

- `orderConfirm: OrderResponse | null`
  - if not null, open modal
- `errorMsg: string | null`
  - shown in error snackbar
- `promoError: string | null`
  - shown under promo code field
- `successMsg: string | null`
  - shown in success snackbar
- `syncWarnings: string[]`
  - shown in warning snackbar

Zustand store state/actions:

- `items`
- `clearCart`
- `syncPrices`

### 8.2 Fetching products (React Query `useQuery`)

```ts
useQuery({ queryKey: ["products"], queryFn: getProducts });
```

Concepts:

- `queryKey` uniquely identifies this cache entry.
- `queryFn` is what actually fetches the data.
- React Query returns:
  - `data` (renamed to `products`)
  - `isLoading`
  - `error`

### 8.3 Sync cart with latest products (`useEffect`)

When `products` arrive:

- call `syncPrices(products)`
- if warnings returned, show them in a warning snackbar

This is a classic React pattern:

- `useEffect` runs “after render” when dependencies change.

### 8.4 Placing an order (React Query `useMutation`)

Mutation setup:

- `mutationFn: placeOrder`
- `onMutate`: clear previous messages
- `onSuccess`:
  - set `orderConfirm`
  - `clearCart()`
  - set success snackbar if promo applied
- `onError`:
  - extract backend error message
  - if coupon-related, set `promoError`
  - set `errorMsg`

Concept: **mutation** is “write” operation (POST/PUT/DELETE).

### 8.5 Rendering the UI

The UI is a 2-column layout using MUI Grid:

- Left: product list
- Right: cart sidebar

Then below:

- `OrderModal` (controlled by `orderConfirm`)
- Snackbars (error/success/warnings)

Important beginner concept: **controlled components**

- The `Cart` doesn’t place orders itself.
- It receives an `onConfirm` callback from the page.
- The page decides what happens (mutation).

---

## 9) Major components deep dive

### 9.1 `ProductCard` (`src/components/products/ProductCard.tsx`)

Goal: show a single product, image, category/name/price, and cart controls.

Key concepts:

- Reads `items` from `useCartStore()` to find current quantity.
- Uses `useMediaQuery` to choose the correct image for screen size.
- Conditional rendering:
  - If quantity is 0: show “Add to Cart” button
  - Else: show +/- quantity stepper

Important patterns:

- `const cartItem = items.find(...)`
  - deriving UI state from store
- `onClick={() => addItem(product)}`
  - event handler updates store, which causes UI to re-render

### 9.2 `Cart` (`src/components/cart/Cart.tsx`)

Goal: show cart items, subtotal, promo code input, confirm order button.

Key concepts:

- Reads from store:
  - `items`
  - `removeItem`
  - `getSubtotal`
  - `getTotalItems`

Two UI states:

- **Empty cart**: show placeholder illustration and message
- **Non-empty**:
  - list items
  - show total
  - promo code input
  - confirm order button

Promo code field:

- Controlled by local state `couponCode`
- If `promoError` exists, MUI `TextField` shows it as helper text and error style

Deletion flow:

- Clicking X sets `deleteItemId`
- That opens a confirmation `Dialog`
- “Remove” calls `removeItem(deleteItemId)`

Beginner concepts inside:

- **Controlled input** (`value`, `onChange`)
- **Derived values**: subtotal is computed via store helper
- **Conditional rendering** based on `totalItems`

### 9.3 `OrderModal` (`src/components/cart/OrderModal.tsx`)

Goal: show order confirmation summary in a modal.

Props:

- `open`: boolean
- `order`: `OrderResponse | null`
- `onReset`: callback to start a new order

Important logic:

- `if (!order) return null;`
  - prevents rendering if no order exists

Rendering order items:

- `order.items.map(item => { ... })`
- For each order item, it finds the full product info:

```ts
const p = order.products.find((prod) => prod.id === item.productId);
```

This is necessary because:

- `OrderItem` only has `productId` and `quantity`
- `OrderResponse.products` holds full `Product` objects

It shows:

- thumbnail
- name
- quantity and unit price
- line total
- overall total and discounts

Button:

- “Start New Order” calls `onReset` (page sets `orderConfirm` back to null)

---

## 10) “Big picture” data flow (super important)

### 10.1 Loading products

1. `Home` mounts
2. `useQuery(['products'], getProducts)` runs
3. Backend returns product list
4. `products` becomes available
5. UI re-renders product grid

### 10.2 Adding/removing items

1. User clicks “Add to Cart” in `ProductCard`
2. `addItem(product)` updates Zustand store
3. Components subscribed to store re-render (`ProductCard`, `Cart`)
4. Cart shows updated items and totals
5. Zustand `persist` writes to localStorage

### 10.3 Confirm order

1. User enters promo code (optional)
2. User clicks “Confirm Order” in `Cart`
3. `Cart` calls `onConfirm(couponCode)` (provided by `Home`)
4. `Home` builds `OrderRequest` from cart items
5. `useMutation(placeOrder).mutate(orderRequest)` sends POST request
6. On success:
   - `orderConfirm` is set
   - cart is cleared
   - modal opens
7. On error:
   - `promoError` set if coupon issue
   - `errorMsg` set for snackbar

---

## 11) Implementation TODO checklist (extend / improve)

### A) Environment setup

- [ ] Create `.env.local` in `frontend-challenge/` (if needed)
- [ ] Set `NEXT_PUBLIC_API_URL=http://localhost:8080`
- [ ] Set `NEXT_PUBLIC_API_KEY=...` (if backend requires it)

### B) State + correctness

- [ ] Prevent ordering when cart is empty (disable Confirm Order)
- [ ] Add validation for coupon code (trim whitespace, max length)
- [ ] Show per-item quantity controls inside `Cart` (optional)
- [ ] Consider storing only `productId` + `quantity` in localStorage (re-hydrate product details from server) to avoid stale product data

### C) React Query improvements

- [ ] Add retry/backoff strategy for `getProducts`
- [ ] Add `refetchOnWindowFocus` decision (on/off depending on UX)
- [ ] Invalidate `['products']` after successful order if backend changes stock/pricing

### D) UX improvements

- [ ] Replace `<img>` with Next.js `<Image>` for optimization (requires setting allowed domains and sizes)
- [ ] Add skeleton loaders for product cards
- [ ] Improve accessibility:
  - [ ] Add `aria-label` to icon buttons
  - [ ] Ensure dialogs have appropriate focus management (MUI mostly handles this)

### E) Error handling

- [ ] Normalize API errors in `services/api.ts` (wrap axios errors into a consistent shape)
- [ ] Display more structured error messages (e.g., coupon invalid vs expired vs not applicable)

### F) Testing (optional but recommended)

- [ ] Add unit tests for `useCartStore` actions (add/remove/update/subtotal)
- [ ] Add component tests for `Cart` and `ProductCard`
- [ ] Add an integration test for the order flow

---

## 12) Quick “how to read this code” tips (beginner)

- If something is about **UI and layout** → check `components/`.
- If something is about **network calls** → check `services/api.ts`.
- If something is about **cart behavior** → check `store/useCartStore.ts`.
- If something is about **page composition / wiring** → check `app/page.tsx`.

---

## 13) Glossary (concepts you just used)

- **Component**: reusable UI function (`function ProductCard() { ... }`)
- **Props**: data passed into a component (`<Cart onConfirm={...} />`)
- **State**: data that changes over time (`useState`, Zustand)
- **Derived state**: calculated values (`subtotal`, `totalItems`)
- **Effect**: logic that runs after render (`useEffect`)
- **Server state**: data from backend (React Query)
- **Client state**: local UI/app state (Zustand)
- **Controlled input**: input whose value is stored in state (`TextField value/onChange`)
- **Mutation**: write operation (POST order)

---
