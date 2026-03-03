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
import { getProducts, placeOrder } from '@/services/api';
import ProductCard from '@/components/products/ProductCard';
import Cart from '@/components/cart/Cart';
import OrderModal from '@/components/cart/OrderModal';
import { useCartStore } from '@/store/useCartStore';
import { OrderResponse } from '@/types';

export default function Home() {
  const { items, clearCart } = useCartStore();
  const [orderConfirm, setOrderConfirm] = React.useState<OrderResponse | null>(null);
  const [errorMsg, setErrorMsg] = React.useState<string | null>(null);
  const [promoError, setPromoError] = React.useState<string | null>(null);
  const [successMsg, setSuccessMsg] = React.useState<string | null>(null);

  // 1. Fetch products
  const { data: products, isLoading, error } = useQuery({
    queryKey: ['products'],
    queryFn: getProducts,
  });

  // 2. Order mutation
  const orderMutation = useMutation({
    mutationFn: placeOrder,
    onMutate: () => {
      setPromoError(null);
      setErrorMsg(null);
      setSuccessMsg(null);
    },
    onSuccess: (data) => {
      setOrderConfirm(data);
      if (data.couponCode) {
        setSuccessMsg(`Promo code ${data.couponCode} applied successfully!`);
      }
    },
    onError: (err: any) => {
      const msg = err.response?.data?.message || err.message || 'Failed to place order';
      if (msg.toLowerCase().includes('coupon code')) {
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
      couponCode,
      items: orderItems
    });
  };

  const handleReset = () => {
    clearCart();
    setOrderConfirm(null);
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
        <Alert severity="error">Error loading products. Please ensure the backend is running at http://localhost:8080</Alert>
      </Container>
    );
  }

  return (
    <Box sx={{ bgcolor: 'background.default', minHeight: '100vh', py: { xs: 4, md: 8 } }}>
      <Container maxWidth="lg">
        <Grid container spacing={4}>
          {/* Product List */}
          <Grid size={{ xs: 12, md: 8 }}>
            <Typography variant="h1" sx={{ mb: 4 }}>
              Desserts
            </Typography>
            <Grid container spacing={3}>
              {products?.map((product) => (
                <Grid key={product.id} size={{ xs: 12, sm: 6, lg: 4 }}>
                  <ProductCard product={product} />
                </Grid>
              ))}
            </Grid>
          </Grid>

          {/* Cart Sidebar */}
          <Grid size={{ xs: 12, md: 4 }}>
            <Box sx={{ position: { md: 'sticky' }, top: 32 }}>
              <Cart onConfirm={handleConfirmOrder} promoError={promoError} />
            </Box>
          </Grid>
        </Grid>

        {/* Confirmation Modal */}
        <OrderModal
          open={!!orderConfirm}
          order={orderConfirm}
          onReset={handleReset}
        />

        {/* Error Snackbar */}
        <Snackbar
          open={!!errorMsg}
          autoHideDuration={6000}
          onClose={() => setErrorMsg(null)}
        >
          <Alert onClose={() => setErrorMsg(null)} severity="error" sx={{ width: '100%' }}>
            {errorMsg}
          </Alert>
        </Snackbar>

        {/* Success Snackbar */}
        <Snackbar
          open={!!successMsg}
          autoHideDuration={6000}
          onClose={() => setSuccessMsg(null)}
        >
          <Alert onClose={() => setSuccessMsg(null)} severity="success" sx={{ width: '100%' }}>
            {successMsg}
          </Alert>
        </Snackbar>
      </Container>
    </Box>
  );
}
