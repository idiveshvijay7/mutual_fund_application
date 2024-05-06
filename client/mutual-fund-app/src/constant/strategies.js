const strategies = [
    {
        "name": "Arbitrage Strategy",
        "description": "This strategy is based on the concept of arbitrage. It is a trading strategy that profits by exploiting the price differences of identical or similar financial instruments on different markets or in different forms. This strategy is based on the concept of arbitrage. It is a trading strategy that profits by exploiting the price differences of identical or similar financial instruments on different markets or in different forms.",
        "funds": [
            {
                "name": "Arbitrage Fund 1",
                "percentage": 10
            },
            {
                "name": "Arbitrage Fund 2",
                "percentage": 20
            },
            {
                "name": "Arbitrage Fund 3",
                "percentage": 30
            },
            {
                "name": "Arbitrage Fund 4",
                "percentage": 40
            }
        ]
    },
    {
        "name": "Balanced Strategy",
        "description": "This strategy is based on the concept of balanced portfolio. It is a trading strategy that profits by exploiting the price differences of identical or similar financial instruments on different markets or in different forms. This strategy is based on the concept of balanced portfolio. It is a trading strategy that profits by exploiting the price differences of identical or similar financial instruments on different markets or in different forms.",
        "funds": [
            {
                "name": "Balanced Fund 1",
                "percentage": 20
            },
            {
                "name": "Balanced Fund 2",
                "percentage": 20
            },
            {
                "name": "Balanced Fund 3",
                "percentage": 5
            },
            {
                "name": "Balanced Fund 4",
                "percentage": 40
            },
            {
                "name": "Balanced Fund 5",
                "percentage": 15
            }
        ]
    },
    {
        "name": "Growth Strategy",
        "description": "This strategy is based on the concept of growth portfolio. It is a trading strategy that profits by exploiting the price differences of identical or similar financial instruments on different markets or in different forms. This strategy is based on the concept of growth portfolio. It is a trading strategy that profits by exploiting the price differences of identical or similar financial instruments on different markets or in different forms.",
        "funds": [
            {
                "name": "Growth Fund 1",
                "percentage": 50
            },
            {
                "name": "Growth Fund 2",
                "percentage": 10
            },
            {
                "name": "Growth Fund 3",
                "percentage": 10
            },
            {
                "name": "Growth Fund 4",
                "percentage": 15
            }, 
            {
                "name": "Growth Fund 5",
                "percentage": 15
            }
        ]
    }
];

// Function to transform fund data into strategy-wise format
export function transformToStrategyWise(fundData) {
    const transformedStrategies = [];

    strategies.forEach(strategy => {
        const transformedStrategy = {
            name: strategy.name,
            funds: [],
            investedAmount: 0,
            marketValue: 0
        };

        strategy.funds.forEach(fund => {
            if (fundData[fund.name]) {
                const fundInfo = {
                    name: fund.name,
                    totalInvest: fundData[fund.name].total_amount,
                    marketValue: fundData[fund.name].market_value
                };

                transformedStrategy.funds.push(fundInfo);
                transformedStrategy.investedAmount += fundData[fund.name].total_amount;
                transformedStrategy.marketValue += fundData[fund.name].market_value;
            }
        });

        transformedStrategies.push(transformedStrategy);
    });

    return transformedStrategies;
}

export default strategies
