// HoldingPage.js
import React from 'react';
import { useLocation, Link } from 'react-router-dom';

const HoldingPage = () => {
    const location = useLocation();
    const { funds } = location.state;

    return (
        <div style={styles.container}>
            <Link to="/home" style={styles.backLink}>Back</Link>
            <h2 style={styles.heading}>Holding Details</h2>
            <table style={styles.table}>
                <thead>
                    <tr>
                        <th style={styles.tableHeader}>Fund Name</th>
                        <th style={styles.tableHeader}>Investment</th>
                        <th style={styles.tableHeader}>Market Value</th>
                    </tr>
                </thead>
                <tbody>
                    {funds.map((fund, index) => (
                        <tr key={index}>
                            <td style={styles.tableCell}>{fund.name}</td>
                            <td style={styles.tableCell}>{fund.totalInvest}</td>
                            <td style={styles.tableCell}>{fund.marketValue}</td>
                        </tr>
                    ))}
                </tbody>
            </table>
        </div>
    );
};

export default HoldingPage;

const styles = {
    container: {
        padding: '20px',
        maxWidth: '600px',
        margin: '0 auto'
    },
    backLink: {
        position: 'absolute',
        top: '10px',
        left: '10px',
        color: '#007bff',
        textDecoration: 'none',
        fontWeight: 'bold'
    },
    heading: {
        textAlign: 'center',
        marginBottom: '20px',
        color: '#333'
    },
    table: {
        width: '100%',
        borderCollapse: 'collapse'
    },
    tableHeader: {
        backgroundColor: '#007bff',
        color: '#fff',
        padding: '10px',
        textAlign: 'left'
    },
    tableCell: {
        border: '1px solid #ccc',
        padding: '10px'
    }
};
