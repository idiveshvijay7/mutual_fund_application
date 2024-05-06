import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { transformToStrategyWise } from '../constant/strategies'

// Component for displaying holding strategy
const HoldingStrategy = ({ strategy, funds, investedAmount, marketValue }) => {
    const navigate = useNavigate();

    const handleSeeHoldings = () => {
        // Navigate to HoldingPage and pass the funds data
        navigate('/holdings', { state: { funds } });
    };

    return (
        <div style={styles.strategyContainer}>
            <div style={styles.headerRow}>
                <h2>{strategy}</h2>
            </div>
            <div style={styles.infoRow}>
                <p>Invested Amount: {investedAmount}</p>
                <p>Market Value: {marketValue}</p>
            </div>
            <div style={styles.actionsRow}>
                <button style={styles.actionButton} onClick={handleSeeHoldings}>See Holdings</button>
            </div>
        </div>
    );
};

// Component to display user's information and holding strategies
const HomePage = () => {
    const [userPhoneNumber, setUserPhoneNumber] = useState('');
    const [holdingStrategies, setHoldingStrategies] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error] = useState(null);
    const navigate = useNavigate(); // Use navigate hook

    useEffect(() => {
        // Fetch user information (phone number) from backend
        // You can replace this with your actual backend API call to get user information
        // For demo purpose, setting a default user phone number
        const phoneNumber = localStorage.getItem('phoneNumber');
        setUserPhoneNumber(phoneNumber);

        // Fetch aggregated orders by phone number
        fetch(`http://localhost:8081/aggregated-orders-by-phone?phoneNumber=${phoneNumber}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({})
        })
        .then(response => {
            if (!response.ok) {
                throw new Error('Error fetching aggregated orders');
            }
            return response.json();
        })
        .then(data => {
            // Transform API data into holding strategies format
            const transformedData = transformToStrategyWise(data);
            // Update holding strategies state with the transformed data
            setHoldingStrategies(transformedData);
        })
        .catch(error => {
            console.error('Error fetching aggregated orders:', error);
            // alert('Error fetching aggregated orders. Please try again later.');
        })
        .finally(() => {
            setLoading(false);
        });
    }, []);

    // Function to handle logout
    const handleLogout = () => {
        // Clear user data from local storage
        localStorage.removeItem('phoneNumber');
        // Navigate to the login page
        navigate('/');
    };

    // Function to navigate to Select Strategy page
    const handleSelectStrategy = () => {
        navigate('/select-strategy');
    };

    if (loading) {
        return <div>Loading...</div>;
    }

    if (error) {
        return <div>Error: {error}</div>;
    }

    return (
        <div>
            <div style={styles.userInfo}>
                <p>Hi, {userPhoneNumber}</p>
                {/* Render logout button if user is authenticated */}
                {userPhoneNumber && (
                    <button style={styles.logoutButton} onClick={handleLogout}>Logout</button>
                )}
            </div>
            <div style={styles.holdingStrategiesContainer}>
                {holdingStrategies.map((strategy, index) => (
                    <HoldingStrategy
                        key={index}
                        strategy={strategy.name}
                        funds={strategy.funds}
                        investedAmount={strategy.investedAmount}
                        marketValue={strategy.marketValue}
                    />
                ))}
            </div>
            {/* Button to navigate to Select Strategy page */}
            <button style={styles.selectButton} onClick={handleSelectStrategy}>Select Strategy For  Transit</button>
        </div>
    );
};

export default HomePage;

const styles = {
    userInfo: {
        textAlign: 'right',
        padding: '20px',
        fontSize: '18px',
        color: '#555'
    },
    holdingStrategiesContainer: {
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center'
    },
    strategyContainer: {
        border: '1px solid #ccc',
        padding: '20px',
        margin: '20px',
        borderRadius: '5px',
        width: '300px',
        textAlign: 'center'
    },
    headerRow: {
        marginBottom: '10px'
    },
    infoRow: {
        marginBottom: '10px'
    },
    actionsRow: {
        marginBottom: '10px'
    },
    actionButton: {
        backgroundColor: '#007bff',
        color: '#fff',
        border: 'none',
        borderRadius: '5px',
        padding: '5px 10px',
        marginBottom: '5px',
        cursor: 'pointer',
        marginRight: '10px'
    },
    logoutButton: {
        backgroundColor: '#dc3545',
        color: '#fff',
        border: 'none',
        borderRadius: '5px',
        padding: '5px 10px',
        marginBottom: '5px',
        cursor: 'pointer',
        marginRight: '10px'
    },
    selectButton: {
        position: 'fixed',
        bottom: '20px',
        right: '20px',
        backgroundColor: '#007bff',
        color: '#fff',
        border: 'none',
        borderRadius: '5px',
        padding: '10px 20px',
        cursor: 'pointer'
    }
};
