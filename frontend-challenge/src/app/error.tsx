'use client';

import { Box, Container, Typography, Button } from '@mui/material';

export default function Error({
    error,
    reset,
}: {
    error: Error & { digest?: string };
    reset: () => void;
}) {
    return (
        <Container maxWidth="sm">
            <Box sx={{ textAlign: 'center', py: 12 }}>
                <Typography variant="h2" sx={{ mb: 2 }}>
                    Something went wrong
                </Typography>
                <Typography variant="body1" color="text.secondary" sx={{ mb: 4 }}>
                    {error.message || 'An unexpected error occurred.'}
                </Typography>
                <Button variant="contained" onClick={reset}>
                    Try again
                </Button>
            </Box>
        </Container>
    );
}
