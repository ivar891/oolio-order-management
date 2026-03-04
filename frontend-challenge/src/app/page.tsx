'use client';

import React from 'react';
import {
  Box,
  Container,
  Typography,
  CircularProgress,
  Alert,
  Snackbar,
  Grid
} from '@mui/material';
import { useQuery, useMutation } from '@tanstack/react-query';
import { AxiosError } from 'axios';
import { getProducts, placeOrder } from '@/services/api';
import ProductCard from '@/components/products/ProductCard';
import Cart from '@/components/cart/Cart';
import OrderModal from '@/components/cart/OrderModal';
import { useCartStore } from '@/store/useCartStore';
import { Product, OrderResponse } from '@/types';

interface APIErrorResponse {
  code: string;
  message: string;
}

function ProductGrid({ products }: { products: Product[] }) {
  return (
    <Grid container spacing={3}>
      {products.map((product) => (
        <Grid key={product.id} size={{ xs: 12, sm: 6, lg: 4 }}>
          <ProductCard product={product} />
        </Grid>
      ))}
    </Grid>
  );
}

function Notifications({
  errorMsg,
  successMsg,
  syncWarnings,
  onClearError,
  onClearSuccess,
  onClearWarnings,
}: {
  errorMsg: string | null;
  successMsg: string | null;
  syncWarnings: string[];
  onClearError: () => void;
  onClearSuccess: () => void;
  onClearWarnings: () => void;
}) {
  return (
    <>
      <Snackbar open={!!errorMsg} autoHideDuration={6000} onClose={onClearError}>
        <Alert onClose={onClearError} severity="error" sx={{ width: '100%' }}>
          {errorMsg}
        </Alert>
      </Snackbar>

      <Snackbar open={!!successMsg} autoHideDuration={6000} onClose={onClearSuccess}>
        <Alert onClose={onClearSuccess} severity="success" sx={{ width: '100%' }}>
          {successMsg}
        </Alert>
      </Snackbar>

      <Snackbar
        open={syncWarnings.length > 0}
        onClose={onClearWarnings}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'left' }}
      >
        <Alert onClose={onClearWarnings} severity="warning" sx={{ width: '100%' }}>
          {syncWarnings.map((msg, idx) => (
            <Box key={idx} sx={{ display: 'block', mb: syncWarnings.length > 1 ? 0.5 : 0 }}>
              {msg}
            </Box>
          ))}
        </Alert>
      </Snackbar>
    </>
  );
}

export default function Home() {
  const items = useCartStore((state) => state.items);
  const clearCart = useCartStore((state) => state.clearCart);
  const syncPrices = useCartStore((state) => state.syncPrices);

  const [orderConfirm, setOrderConfirm] = React.useState<OrderResponse | null>(null);
  const [errorMsg, setErrorMsg] = React.useState<string | null>(null);
  const [promoError, setPromoError] = React.useState<string | null>(null);
  const [successMsg, setSuccessMsg] = React.useState<string | null>(null);
  const [syncWarnings, setSyncWarnings] = React.useState<string[]>([]);

  const { data: products, isLoading, error } = useQuery({
    queryKey: ['products'],
    queryFn: getProducts,
  });

  React.useEffect(() => {
    if (products?.length) {
      const msgs = syncPrices(products);
      if (msgs.length > 0) {
        setSyncWarnings(msgs);
      }
    }
  }, [products, syncPrices]);

  const orderMutation = useMutation({
    mutationFn: placeOrder,
    onMutate: () => {
      setPromoError(null);
      setErrorMsg(null);
      setSuccessMsg(null);
    },
    onSuccess: (data) => {
      setOrderConfirm(data);
      clearCart();
      if (data.couponCode) {
        setSuccessMsg(`Promo code ${data.couponCode} applied successfully!`);
      }
    },
    onError: (err: AxiosError<APIErrorResponse>) => {
      const apiError = err.response?.data;
      const msg = apiError?.message || err.message || 'Failed to place order';

      if (apiError?.code === 'validation' || apiError?.code === 'bad_request') {
        setPromoError(msg);
      }
      setErrorMsg(msg);
    }
  });

  const handleConfirmOrder = (couponCode: string) => {
    const orderItems = items.map(item => ({
      productId: item.productId,
      quantity: item.quantity
    }));

    orderMutation.mutate({
      couponCode: couponCode.trim() || undefined,
      items: orderItems
    });
  };

  if (isLoading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    return (
      <Container maxWidth="lg" sx={{ py: 8 }}>
        <Alert severity="error">
          Unable to load products. Please check your connection and try again.
        </Alert>
      </Container>
    );
  }

  return (
    <Box sx={{ bgcolor: 'background.default', minHeight: '100vh', py: { xs: 4, md: 8 } }}>
      <Container maxWidth="lg">
        <Grid container spacing={4}>
          <Grid size={{ xs: 12, md: 8 }}>
            <Typography variant="h1" sx={{ mb: 4 }}>
              Desserts
            </Typography>
            {products && <ProductGrid products={products} />}
          </Grid>

          <Grid size={{ xs: 12, md: 4 }}>
            <Box sx={{ position: { md: 'sticky' }, top: 32 }}>
              <Cart
                onConfirm={handleConfirmOrder}
                promoError={promoError}
                onPromoErrorClear={() => setPromoError(null)}
                isPending={orderMutation.isPending}
              />
            </Box>
          </Grid>
        </Grid>

        <OrderModal
          open={!!orderConfirm}
          order={orderConfirm}
          onReset={() => setOrderConfirm(null)}
        />

        <Notifications
          errorMsg={errorMsg}
          successMsg={successMsg}
          syncWarnings={syncWarnings}
          onClearError={() => setErrorMsg(null)}
          onClearSuccess={() => setSuccessMsg(null)}
          onClearWarnings={() => setSyncWarnings([])}
        />
      </Container>
    </Box>
  );
}
