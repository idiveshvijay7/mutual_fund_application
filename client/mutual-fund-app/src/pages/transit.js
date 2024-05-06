// TransitPage.js
import React, { useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import axios from 'axios';

const TransitPage = () => {
    const location = useLocation();
    const [investmentAmount, setInvestmentAmount] = useState('');
    const { strategy } = location.state;
    const navigate = useNavigate();

    // Function to handle investment

    const handleInvest = () => {
        if (!investmentAmount || investmentAmount <= 0) {
            alert('Please enter a valid investment amount.');
            return;
        }
    
        // Construct the request body
        const requestBody = {
            accountNumber: '11200222',
            ifscCode: 'UBIT22222',
            amount: parseInt(investmentAmount),
            redirectUrl: 'http://localhost:3000',
            strategyName: strategy.name
        };
    
        // Make a POST request to the payment endpoint
        axios.post('http://localhost:8080/payment', requestBody)
            .then(response => {
                console.log('Response data:', response.data);
                const { paymentLink } = response.data;
                const completeURL = paymentLink;
                window.location.href = completeURL;
            })
            .catch(error => {
                console.error('Error:', error);
                if (error.response) {
                    // The request was made and the server responded with a status code
                    // that falls out of the range of 2xx
                    alert('Payment failed. Please try again later.');
                    console.error('Server response:', error.response.data);
                    console.error('Server status code:', error.response.status);
                } else if (error.request) {
                    // The request was made but no response was received
                    alert('Payment failed. Please check your internet connection and try again.');
                    console.error('No response received:', error.request);
                } else {
                    // Something happened in setting up the request that triggered an Error
                    alert('Payment failed due to an unexpected error. Please try again later.');
                    console.error('Request setup error:', error.message);
                }
            });
    };
    


    // Function to calculate allocated amount for each fund based on percentage
    const calculateAllocatedAmount = (percentage) => {
        const totalPercentage = strategy.funds.reduce((acc, fund) => acc + fund.percentage, 0);
        const amountPerPercentage = investmentAmount / totalPercentage;
        return (amountPerPercentage * percentage).toFixed(2); // Limit decimal places to 2
    };

    // Function to handle going back
    const handleGoBack = () => {
        navigate(-1); // Go back to the previous page
    };

    // Function to handle input change and allow only numbers
    const handleInputChange = (e) => {
        const value = e.target.value;
        if (/^\d*$/.test(value)) { // Only allow numbers
            setInvestmentAmount(value);
        }
    };

    return (
        <div style={styles.container}>
            <h1 style={styles.heading}>Enter Investment Amount</h1>
            <button onClick={handleGoBack} style={styles.backButton}>Back</button>
            <div style={styles.inputContainer}>
                <input
                    type="text"
                    value={investmentAmount}
                    onChange={handleInputChange}
                    placeholder="Enter amount"
                    style={styles.input}
                />
            </div>
            <div style={styles.strategyDetails}>
                <h2 style={styles.strategyName}>{strategy.name}</h2>
                <div style={styles.fundsContainer}>
                    {strategy.funds.map((fund, index) => (
                        <div key={index} style={styles.fundItem}>
                            <p style={styles.fundName}>{fund.name}</p>
                            <div style={styles.amountBox}>{calculateAllocatedAmount(fund.percentage)}</div>
                        </div>
                    ))}
                </div>
            </div>

            <button onClick={handleInvest} style={styles.investButton}>Invest</button>
        </div>
    );
};


export default TransitPage;

const styles = {
    container: {
        textAlign: 'center',
        marginTop: '50px',
        fontFamily: 'Arial, sans-serif',
    },
    heading: {
        fontSize: '24px',
        marginBottom: '20px',
        color: '#333',
    },
    backButton: {
        backgroundColor: '#007bff',
        color: '#fff',
        border: 'none',
        borderRadius: '5px',
        padding: '8px 16px',
        cursor: 'pointer',
        position: 'absolute',
        top: '20px',
        left: '20px',
    },
    inputContainer: {
        marginBottom: '20px',
    },
    input: {
        padding: '8px',
        fontSize: '16px',
        width: '200px',
        borderRadius: '5px',
        border: '1px solid #ccc',
    },
    strategyDetails: {
        marginBottom: '20px',
        textAlign: 'center',
    },
    strategyName: {
        fontSize: '20px',
        fontWeight: 'bold',
        color: '#007bff',
        marginBottom: '10px',
    },
    fundsContainer: {
        marginTop: '20px',
    },
    fundItem: {
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        marginBottom: '10px',
        width: '100%',
    },
    fundName: {
        fontWeight: 'bold',
        width: '50%',
        textAlign: 'center',
    },
    allocatedAmount: {
        color: '#007bff',
        width: '50%',
        textAlign: 'center',
    },
    investButton: {
        backgroundColor: '#007bff',
        color: '#fff',
        border: 'none',
        borderRadius: '5px',
        padding: '10px 20px',
        fontSize: '16px',
        cursor: 'pointer',
        marginTop: '20px',
    },
    amountBox: {
        width: '100px',
        height: '40px',
        backgroundColor: '#f0f0f0',
        borderRadius: '5px',
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        margin: '0 auto', // Center the box horizontally
        fontSize: '16px',
        fontWeight: 'bold',
        color: '#333',
        marginBottom: '10px',
    },
};
