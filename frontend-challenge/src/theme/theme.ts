import { createTheme } from '@mui/material/styles';

export const getTheme = (fontFamily: string) =>
    createTheme({
        palette: {
            primary: {
                main: '#C73B0F', // Rose 600 from design (approx)
                contrastText: '#FFFFFF',
            },
            secondary: {
                main: '#261108', // Rose 900
            },
            background: {
                default: '#F5F5F5',
                paper: '#FFFFFF',
            },
            text: {
                primary: '#261108',
                secondary: '#87635A',
            },
        },
        typography: {
            fontFamily: fontFamily,
            h1: {
                fontSize: '40px',
                fontWeight: 700,
                lineHeight: '120%',
            },
            h2: {
                fontSize: '24px',
                fontWeight: 700,
            },
            body1: {
                fontSize: '16px',
                fontWeight: 400,
            },
            body2: {
                fontSize: '14px',
                fontWeight: 400,
            },
        },
        shape: {
            borderRadius: 8,
        },
        components: {
            MuiButton: {
                styleOverrides: {
                    root: {
                        textTransform: 'none',
                        borderRadius: '999px',
                        fontWeight: 600,
                        padding: '12px 24px',
                    },
                    containedPrimary: {
                        '&:hover': {
                            backgroundColor: '#A82F0A',
                        },
                    },
                },
            },
        },
    });
