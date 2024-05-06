# Mutual Fund application

Aim of this assignment is to create a mutual fund application.

## Requirments

- List the Mutual strategies
  - Each strategy will have a list of mutual funds
    - Data will added along with assignment
  - Example: Say Strategy 1 will have 5 mutual funds with 20%, 50%, 10%, 5%, 15% allocation. So if a customer invests Rs. 1L, the investments has to be made for
    | Fund Name | Amount |
    | --------- | ------ |
    | Fund 1 | 20K |
    | Fund 2 | 50K |
    | Fund 3 | 10K |
    | Fund 4 | 5K |
    | Fund 5 | 15K |
- Customer should be able to select the strategy.
- Enter the amount
- Make payment
  - Payment gateway is provided along with this assignment.
  - Follow the documentation of payment gateway to set it up.
- Record the payment details
- Push the order to RTA
  - There is a RTA service added part of this assignment.
  - RTA also have apis to fetch the status.
  - RTA service setup information is provided.
- User should be able to see his portfolio details as and when order is processed.
  - Portfolio will contain the following details
  - List of Strategies
    - Invested amount per strategy
    - Current market value of the strategy. (RTA service has an api to get the current market value of a fund)
    - List of funds in the strategy
      - Invested amount per fund
      - Current market value of the funds.

## Deliverables

1. A web application with UI and backend that can facilitate purchase of mutual fund through strategies. No need of strict authentication. Can keep user's phone in request headers for authentication. Phone can be the primary id to keep things simple.
2. User should be able to login by giving phone number
3. See the strategies
4. Make a payment
5. After payment see the portfolio in some time

Basic wireframe for the UI design reference can be found at [Figma link](https://www.figma.com/file/y1vgJNCPvOH2yyEKGCHDDu/Mutual-Fund-App-Assignment?type=design&node-id=0%3A1&mode=design&t=kBBvBLZqrhi6NFM9-1).

## Evaluation criteria

- Code should clean, maintainable and readable
  - Try to avoid nested if, nested for loop etc.
  - Handle the errors gracefully
- There should be proper readme file to setup the project. Just by following the readme one should be able to run the application.
- We recommend wrapping everything inside a docker compose or any container runtime orchestrator of your choice. This is to ensure that it works in my local and doesn't work in your system doesn't happen.
- You may or may not choose a code generator. But should be aware that these tools may not handle the errors properly or write the code in the way you want. So we leave it up to your due deligence.
- There will be a pair programming round based on the assignment if evaluation criteria passes.
- Deliverables should be cleanly met.
- Payment gateway and RTA Service will have glitches in their functioning. It will be explained in their readme. But the platform should be built considering those and try to be resiliant.
- We will run automate test suite to test on your assignment. It should have decent pass score.

## Backend

- Have attached the doc with requests and expected json responses along with the problem.
- Stick with that api schema. Cause our automation test suite will hit assuming the same.
- You can choose whatever database you want. If you are choosing to use a managed service, make sure your db is clean before submitting the assignment so that automated test suite works properly.
