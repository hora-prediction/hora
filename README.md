# Hora: Hierarchical Online Failure Prediction [![Build Status](https://travis-ci.org/hora-prediction/hora.svg?branch=master)](https://travis-ci.org/hora-prediction/hora)

Complex software systems experience failures at runtime even though a lot of effort is put into the development and operation. Reactive approaches detect these failures after they have occurred and already caused serious consequences. In order to execute proactive actions, the goal of online failure prediction is to detect these failures in advance by monitoring the quality of service or the system events. Current failure prediction approaches look at the system or individual components as a monolith without considering the architecture of the system. They disregard the fact that the failure in one component can propagate through the system and cause problems in other components.

Hora is a hierarchical online failure prediction approach which combines component failure predictors with architectural knowledge. The failure propagation is modeled using Bayesian networks which incorporate both prediction results and component dependencies extracted from the architectural models.

Publication: http://www.sciencedirect.com/science/article/pii/S0164121217300390
