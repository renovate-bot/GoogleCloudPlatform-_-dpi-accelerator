# BECKN-ONIX
This project contains the core components for setting up a Beckn-compliant network, including the BAP Adapter, BPP Adapter, Registry, and Gateway. It provides a foundational framework to facilitate seamless interaction and data exchange within the Beckn Protocol ecosystem.


## Beckn Components
The Beckn Protocol is an open-source, open-network protocol designed to enable the discovery and fulfillment of services in a decentralized manner. This project implements key components necessary to participate in a Beckn network.

1. BAP Adapter
   The BAP (Buyer App) Adapter acts as an intermediary for Buyer Applications. It is responsible for:

   * Translating requests from a Buyer Application into Beckn-compliant messages.
   * Routing these messages to the appropriate Gateway.
   * Receiving responses from the network and translating them back for the Buyer App.
   * Essentially, it allows any application to seamlessly integrate with the Beckn network as a BAP.

1. BPP Adapter
   The BPP (Seller App) Adapter serves as the interface for Seller Applications. Its primary functions include:

   * Receiving Beckn-compliant requests (e.g., search, order) from the network via the Gateway.
   * Translating these requests into a format compliant as per the Seller App.
   * Sending responses from the Seller Application back to the network in a Beckn-compliant manner.
   * This component enables seller-side applications to expose their services and products on the Beckn
   network.

1. Registry
   The Registry is a crucial component that maintains a decentralized database of all participants (BAPs, BPPs, Gateways) in a Beckn network. Its key roles are:

   * Storing public keys and network addresses of registered entities.
   * Enabling discovery of network participants by other components.
   * Ensuring the authenticity and security of communication within the network through
   cryptographic verification.

1. Gateway
   The Gateway acts as a central routing point within a Beckn network, facilitating communication between different participants. Its responsibilities include:

   * Receiving requests from BAP Adapters and routing them to the appropriate BPP Adapters based on
   information from the Registry.
   * Handling message validation and security checks.
   * Ensuring efficient and reliable message delivery across the network.



## Architecture Diagram


## Installation and Configuration
This section details how to get the beckn-onix components up and running.

### Prerequisities

### Configuration
Each Onix component (Gateway, Registry, Adapters) is configured using its own YAML file. These files allow you to set up server ports, logging levels, and service-specific parameters like database connections, timeouts and cache settings etc.

For a detailed explanation of all available configuration values for each component, please refer to the [Onix Configuration README](./onix/configs/README.md).


### Installing for Existing Networks


### Installing for New Networks


## Local Development (Customizing and Building)


## Licensing

* See [LICENSE](LICENSE)


