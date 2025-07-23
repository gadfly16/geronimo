# Introduction

Geronimo is a personal asset rebalancer application that can connect to your cryptocurrency exchange accounts and autonomously trade on markets on your behalf. While most crypto trading tools are about gaining an edge over the market, Geronimo looks at trading as a community effort where individual traders are providing liquidity concurrently to the ecosystem while earning a consistent return in exchange. Geronimo only implements one strategy called Stochastic Proportional Rebalancer (SPR) that is carefully designed for this exact use case. Geronimo is the end result of more than ten years of thinking about cryptocurrencies and experimenting with mining, staking and trading them, so understanding why it is the way it is might need some effort and patience.

## Rebalancing

Rebalancing is a trading strategy where you define a price range with a high and a low limit. The strategy automatically adjust your asset allocation based on where the current price falls within that range. At the high limit, you hold all quote currency (like USD). At the low limit, you hold all base currency (like BTC). At any price between these limits, the algorithm calculates the ideal proportion of each asset you should hold. The strategy generates profits by only executing trades when your actual holdings deviate from this ideal proportion by more than a threshold you specified.

In practice you choose a market that you want to rebalance and the algorithm continuously sells and buys portions of your funds to keep your funds in the right proportion based on your settings. If the price goes up it sells a little and when the price goes down it buys a little constantly realizing and accumulating the partial gain opportunities present in the price movement.

One useful way to look at rebalancing is that it's like mining but instead of getting your reward for providing security, you get it for providing liquidity to the markets and reducing volatility.

It's worth mentioning that the value of the funds managed by a rebalancer is less exposed to the market price since normally your funds are a mixture of the two traded currencies. Basically with it you can get lower price exposure for giving up possible price gains. Rebalancing not only reduces volatility on the market by a tiny bit in the long run, but reduces the volatility of your personal investments instantly.

## Risks

The most important risk that you have to assess when considering rebalancing is that you may end up with a large amount of one asset of the pair that you are rebalancing wile none of the other. If you consider either case *catastrophic* then please don't rebalance that pair.

The trick that can make you sleep well despite the fact that you're trading crypto is to be able to feel good about every price movement. For example I'm personally rebalancing the ADA/EUR pair the most because when the price moves up I'm happy because I can realize my gains and when the price goes down I'm happy because I can buy ADA for cheap. If a black swan event moves the price out of the range I set I might end up with a bunch of ADA which might be scary but not catastrophic since I'm already an investor so that risks is already swallowed or I can end up with a bunch of Euros that is less than I could have made with hodling the same funds, but not catastrophic as it is usually considered a good outcome.

## Costs

SPR is a low frequency, low overhead algorithm that doesn't consume considerable amounts of electricity or bandwidth. The execution cost of trades on the exchange is the only real associated cost of rebalancing itself. Keep in mind that for a complete cost estimation for a rebalancing venture you need to calculate possible deposit and withdrawal costs and taxes and possibly other things.

## The SPR Algorithm

Geronimo concentrates solely on the Stochastic Proportional Rebalancer strategy. Every trader you define in Geronimo runs this algorithm but the individual traders can have different parameter settings so their behavior might differ significantly.

The strategy is founded on the premiss that we can not predict exactly how the price will move short term. In other words at a given moment of time you are not able to tell if the price will be higher or lower the next time you check it. The network of causes moving market prices is just far too complex to enable that. Strategies based on technical and analitycal techniques usually try to answer the question *when* to trade, while rebalancing techniques concentrate on the question *how much* to trade at a given price point. While it would be nice to know *when* it is beneficial to trade, we just dont know that and this makes the question of *when* insignificant.



The SPR algorithm has three *main* parameters: a low limit, a high limit and a threshold.






For example, when the market price is at the geometric mean of your high and low limits, the ideal allocation is 50/50 between the two assets.

## A Personal Rebalancer Application

Geronimo is a software that is designed to run on small *personal servers* (PS) having relatively stable internet connections. A properly configured RasPI 3 connected to the router your internet provider installed for your connection should serve several users without any problems. Geronimo follows a simple server-client architecture where the backend maintains the connections to the markets and serves the GUI clients of users. The back-end is a single executable written in Go, while the primary front-end is a GUI running in browsers.

Geronimo's primary goal is to allow its users to do rebalancing as comfortably and as reliably as possible. Geronimo only supports the SPR algorithm which is designed for this exact purpose. Unfortunately for the time being exchange support is lacking. V1.0 focuses only on Kraken and Binance support, with the adapter architecture designed to make adding new exchanges easy and straightforward.

Geronimo is and always will be free and open source software to let you review, fork and customize it as necessary.

## Security

Instances only handle exchange api-keys thats' permissions can be limited. It's important to appropriately set up the used api-key's permissions to contain the possible consequences of a compromised instance. SPR does not need any withdrawal or deposit privileges, so never allow that for the api keys you use with Geronimo.

## User Experience

After installing Geronimo on your personal server you can immediately connect to it via a browser and start to configure it. Geronimo aims to provide a *set up and forget* user experience since rebalancing doesn't necessarily need any supervision.

Geronimo presents its state through a web interface that updates all displayed information in real-time as they change on the back-end. Geronimo is also constantly displaying the health of the connection between the GUI and the BE so if the connection is healthy you can be sure that the information presented is only a couple of seconds out of date at the worst.

The GUI only connects to the BE and it is not allowed to collect information from external sources on its own. This ensures that what you see is exactly the same as what the BE sees.

The GUI only displays the state of the BE and never any desirable state. For example a parameter changed by the user is only updated after the change has been committed on the BE.

All parameter states are journaled to let the user review previous system states with minimal effort.

## The Tree

In Geronimo every user-manageable item exists in the form of a node. For example if you want to connect to an account on an exchange you need to create an `Account` node and set its parameters up correctly to let the system connect to it. Nodes can have parent-child relationships to allow the entire system to be organized into one single *tree*. Geronimo only supports very few node kinds. As mentioned `Account` nodes connect to exchange accounts and provide *funds* to the tree. `Trader` nodes can *claim* funds provided by `Account` nodes and execute the rebalancing algorithm. `Group` nodes can be used to create groups of other nodes and they can also claim funds that allows easy partitioning.

While ordinary users can only see these node kinds and their dedicated branch on the system tree, users administrator privileges can see the entire tree and all the system node kinds. For example there is `User` node type that describes users and their settings.
