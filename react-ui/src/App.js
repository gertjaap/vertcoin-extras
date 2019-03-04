import React, { Component } from 'react';
import {
    Collapse,
    Navbar,
    NavbarToggler,
    NavbarBrand,
    Nav,
    NavItem,
    NavLink,
    Container,
    Row,
    Col,
    Jumbotron,
    Table,
    Badge,
    Button,
    Form,
    FormGroup,
    Label,
    Input,
    InputGroup,
    InputGroupAddon
} from 'reactstrap';
import {MdRefresh, MdCheckCircle} from 'react-icons/md';
import {BigNumber} from 'bignumber.js';
import QRCode from 'qrcode.react';

class App extends Component {
    constructor(props) {
        super(props);
        this.baseUrl = "/api/"
        if(window.location.host === "localhost:3000") { // When running in REACT Dev
            this.baseUrl = "http://localhost:27888/api/"
        }
        this.toggle = this.toggle.bind(this);
        this.state = {
            isOpen: false,
            balance: new BigNumber(0),
            stealthBalance: new BigNumber(0),
            page: 'home',
            vertcoinAddress: "",
            assetAddress: "",
            assets: [],
            network: "",
            sendAsset: {AssetID:"VTC", Ticker:"VTC", Decimals: 8},
            sendAmount: "0",
            sendTo:"",
            issueTicker: "",
            issueDecimals: 8,
            issueSupply: 0,
            donating: false,
            connected: false,
            synced: false,
            blockHeight: 0,
            headerQueue: 0,
            sendStealth: false,
        };
        this.refreshAssets = this.refreshAssets.bind(this);
        this.refreshBalance = this.refreshBalance.bind(this);
        this.sendAsset = this.sendAsset.bind(this);
        this.getAssetByID = this.getAssetByID.bind(this);
        this.issueAsset = this.issueAsset.bind(this);
        this.refreshNetwork = this.refreshNetwork.bind(this);
        this.refreshAddresses = this.refreshAddresses.bind(this);
        this.refreshSyncStatus = this.refreshSyncStatus.bind(this);
        this.refresh = this.refresh.bind(this);
    }

    sendAsset(asset, amount, to, stealth) {
        var realAmount = new BigNumber(amount).times(new BigNumber("1e" + asset.Decimals.toString())).toNumber();
        fetch(this.baseUrl + "transferAsset", {
            method: "POST",
            headers: {
                'Accept': 'application/json',
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                
                    "AssetID": asset.AssetID,
                    "Amount" : realAmount,
                    "RecipientAddress": to,
                    "UseStealth" : stealth
                
            })
        })
        .then((res) => { return res.json(); })
        .then((data) => {
            this.setState({page:'home'});
        })
    }

    getAssetByID(assetID) {
        var sendAsset = this.state.assets.find((a) => a.AssetID === assetID);
        if (assetID === "VTC") {
            sendAsset = {AssetID: "VTC", Ticker: "VTC", Decimals: 8}
        } else if (assetID === "SVTC") {
            sendAsset = {AssetID: "SVTC", Ticker: "VTC", Decimals: 8}
        }

        return sendAsset;
    }

    issueAsset(ticker, decimals, supply) {
        var realAmount = new BigNumber(supply).times(new BigNumber("1e" + decimals.toString())).toNumber();
        fetch(this.baseUrl + "newAsset", {
            method: "POST",
            headers: {
                'Accept': 'application/json',
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                    "Ticker": ticker,
                    "TotalSupply" : realAmount,
                    "Decimals": parseInt(decimals)
            })
        })
        .then((res) => { return res.json(); })
        .then((data) => {
            this.setState({page:'home'});
        })
    }

    refreshBalance() {
        fetch(this.baseUrl + "balance").then((resp) => resp.json()).then((resp) => {
            this.setState({
                balance: new BigNumber(resp.TotalBalance),
                stealthBalance: new BigNumber(resp.StealthBalance)
            });
        });
    }

    refreshSyncStatus() {
        fetch(this.baseUrl + "syncStatus").then((resp) => resp.json()).then((resp) => {
            this.setState({
                connected: resp.Connected,
                blockHeight: resp.SyncHeight,
                synced: resp.Synced,
                headerQueue: resp.HeaderQueue,
            }, ()=> {
                if(!(this.state.connected === true) && this.state.newRpcServer === undefined) {
                    fetch(this.baseUrl + "rpcSettings").then((resp) => resp.json()).then((resp) => {
                        this.setState({
                            newRpcServer: resp.RpcHost,
                            newRpcUser: resp.RpcUser,
                            newRpcPassword: resp.RpcPassword
                        });
                    });
                }
            });
        });
    }
    refreshAssets() {
        fetch(this.baseUrl + "assets/mine").then((resp) => resp.json()).then((resp) => {
            if(resp.Assets === null) { resp.Assets = []; }
            this.setState({
                assets : resp.Assets
            });
        });
    }
    refreshNetwork() {
        fetch(this.baseUrl + "network").then((resp) => resp.json()).then((resp) => {
            this.setState({
                network : resp.NetworkName
            });
        });
    }
    
    refreshAddresses() {
        fetch(this.baseUrl + "addresses").then((resp) => resp.json()).then((resp) => {
            this.setState({
                vertcoinAddress: resp.VertcoinAddress,
                assetAddress: resp.AssetAddress,
                stealthAddress: resp.StealthAddress
            })
        });
    }

    componentDidMount() {
        this.refreshNetwork();
        this.refresh();
        this.refreshInterval = setInterval(this.refresh, 5000);
    }

    componentWillUnmount() {
        clearInterval(this.refreshInterval);
    }

    refresh() {
        this.refreshBalance();
        this.refreshAddresses();
        this.refreshAssets();
        this.refreshSyncStatus()
    }

    updateRpcSettings() {
        fetch(this.baseUrl + "updateRpcSettings", {
            method: "POST",
            headers: {
                'Accept': 'application/json',
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                
                    "RpcHost": this.state.newRpcServer,
                    "RpcUser" : this.state.newRpcUser,
                    "RpcPassword": this.state.newRpcPassword
                
            })
        })
        .then((res) => { return res.json(); })
        .then((data) => {
            this.setState({connected:true, page:'home'}, () => {this.refresh();});
        })
    }

    toggle() {
        this.setState({
            isOpen: !this.state.isOpen
        });
    }
    render() {
        var page = this.state.page;
        if(!(this.state.connected === true)) {
            page = 'disconnected';
        }
        var fractions = this.state.balance.mod(new BigNumber("1e8"))
        var coins = this.state.balance.plus(fractions.negated()).div(new BigNumber("1e8"))
        
        var stealthFractions = this.state.stealthBalance.mod(new BigNumber("1e8"))
        var stealthCoins = this.state.stealthBalance.plus(stealthFractions.negated()).div(new BigNumber("1e8"))
        

        var mainPage = "";
        switch(page) {
            case 'disconnected':
                mainPage = (<Container>
                <Row>
                    <Col>
                        <h2>Disconnected :-(</h2>
                        <p>Vertcoin OpenAssets is unable to connect to Vertcoin Core. Please check the settings below and correct them if needed:</p>
                        <Form>
                            <FormGroup row>
                                <Label for="amount" sm={2}>RPC URL:</Label>
                                
                                <Col sm={10}>
                                    <InputGroup>
                                        <Input type="text" name="rpcServer" id="rpcServer" placeholder="Enter server URL" value={this.state.newRpcServer} onChange={e => this.setState({ newRpcServer: e.target.value })} />
                                    </InputGroup>
                                </Col>
                            </FormGroup>
                            <FormGroup row>
                                <Label for="amount" sm={2}>RPC User:</Label>
                                
                                <Col sm={10}>
                                    <InputGroup>
                                        <Input type="text" name="rpcUser" id="rpcUser" placeholder="Enter RPC User" value={this.state.newRpcUser} onChange={e => this.setState({ newRpcUser: e.target.value })} />
                                    </InputGroup>
                                </Col>
                            </FormGroup>
                            <FormGroup row>
                                <Label for="amount" sm={2}>RPC Password:</Label>
                                
                                <Col sm={10}>
                                    <InputGroup>
                                        <Input type="text" name="rpcPassword" id="rpcPassword" placeholder="Enter RPC Password" value={this.state.newRpcPassword} onChange={e => this.setState({ newRpcPassword: e.target.value })} />
                                    </InputGroup>
                                </Col>
                            </FormGroup>
                            <FormGroup row>
                                <Col>
                                    <Button onClick={(e) => {
                                        this.updateRpcSettings()
                                    }}>Update</Button>
                                </Col>
                            </FormGroup>
                        </Form>
                    </Col>
                </Row>
                </Container>)
                break;
            default:
            case 'home':
                var assets = this.state.assets.map((asset) => {
                    var assetBalance = new BigNumber(asset.Balance);
                    var divider = new BigNumber("1e" + asset.Decimals.toString());
                    var assetFractions = assetBalance.mod(divider);
                    var assetAmount = assetBalance.plus(assetFractions.negated()).div(divider);
        
                    return (<tr>
                        <td>{asset.AssetID.substring(0,8)}</td>
                        <td>{asset.Ticker}</td>
                        <td><b>{assetAmount.toString()}</b>.<small>{assetFractions.toString()}</small></td>
                        <td>
                            <Button color="primary" size="sm" onClick={(e) => {
                                
                                this.setState({sendAsset:this.getAssetByID(asset.AssetID), page:'send'});
                            }}>
                                Send
                            </Button>
                        </td>
                    </tr>);
                });

                var syncStatus = null;
                if(!this.state.synced) {
                    syncStatus = (<Row>
                        <Col>
                            <h2>Catching up...</h2>
                            <p>Vertcoin OpenAssets is catching up with the blockchain. There's <b>{this.state.headerQueue}</b> blocks left to process</p>
                        </Col>
                    </Row>);
                }

                mainPage = (<Container>
                    {syncStatus}
                    <Row>
                        <Col>
                            <h2>Assets</h2>
                            <Table>
                                <thead>
                                    <tr>
                                        <th>Asset ID</th>
                                        <th>Symbol</th>
                                        <th>Balance</th>
                                        <th>&nbsp;</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {assets}
                                </tbody>
                            </Table>
                        </Col>
                    </Row>
                    <Row>
                        <Col>
                            <h2>Stealth Balance</h2><br/>
                            You have <b>{ stealthCoins.toString() }</b>.<small>{ stealthFractions.toString() }</small>&nbsp;<b>VTC</b> received in Stealth Transactions
                        </Col>
                    </Row>
                </Container>)


                break;
            case 'receive':
                mainPage = (<Container>
                    <Row>
                        <Col xs={6} lg={4} style={{textAlign: "center"}}>
                            <h3>Receive VTC:</h3>
                            <QRCode renderAs="svg" bgColor="rgba(0,0,0,0)" value={this.state.vertcoinAddress} /><br/>
                            <small>{this.state.vertcoinAddress}</small>                        
                        </Col>
                        <Col xs={6} lg={4} style={{textAlign: "center"}}>
                            <h3>Receive Assets:</h3>
                            <QRCode renderAs="svg" bgColor="rgba(0,0,0,0)" value={this.state.assetAddress} /><br/>
                            <small>{this.state.assetAddress}</small>                        
                        </Col>
                        <Col xs={6} lg={4} style={{textAlign: "center"}}>
                            <h3>Stealth Receive VTC:</h3>
                            <QRCode renderAs="svg" bgColor="rgba(0,0,0,0)" value={this.state.stealthAddress} /><br/>
                            <small>{this.state.stealthAddress}</small>                        
                        </Col>
                    </Row>
                </Container>);
                break;
            case 'send':
                var assetOptions = this.state.assets.map((a) => {
                    return (<option value={a.AssetID}>{a.Ticker}</option>)
                });
                mainPage = (<Container>
                    <Row>
                        <Col>
                            <Form>
                                <FormGroup row>
                                    <Label for="asset" sm={2}>Asset:</Label>
                                    <Col sm={10}>
                                    <Input type="select" name="asset" id="asset" value={this.state.sendAsset.AssetID} onChange={e => { console.log(e.target.value); this.setState({ sendAsset: this.getAssetByID(e.target.value) }); }}>
                                            <option value="VTC">VTC</option>
                                            <option value="SVTC">VTC (Stealth)</option>
                                            {assetOptions}
                                    </Input>
                                    </Col>
                                </FormGroup>
                                <FormGroup row>
                                    <Label for="amount" sm={2}>Amount:</Label>
                                    
                                    <Col sm={10}>
                                        <InputGroup>
                                            <Input type="number" name="amount" id="amount" placeholder="Enter amount" value={this.state.sendAmount} onChange={e => this.setState({ sendAmount: e.target.value })} />
                                            <InputGroupAddon addonType="append">{this.state.sendAsset.Ticker}</InputGroupAddon>
                                        </InputGroup>
                                    </Col>
                                </FormGroup>
                                <FormGroup row>
                                    <Label for="recipient" sm={2}>Recipient:</Label>
                                    <Col sm={10}>
                                        <Input type="text" name="recipient" id="recipient" placeholder="Enter recipient address" value={this.state.sendTo} onChange={e => this.setState({ sendTo: e.target.value })} />
                                    </Col>
                                </FormGroup>
                                <FormGroup row style={{display:((this.state.sendAsset.AssetID==="VTC"||this.state.sendAsset.AssetID==="SVTC") ? '' : 'none')}}> 
                                    <Label for="stealth" sm={2}>Use Stealth Inputs:</Label>
                                    <Col sm={10}>
                                        <Input type="checkbox" name="stealth" id="stealth" checked={this.state.sendStealth} onChange={e => this.setState({ sendStealth: e.target.checked })} />
                                    </Col>
                                </FormGroup>
                                <FormGroup row>
                                    <Col>
                                        <Button onClick={(e) => {
                                            this.sendAsset(this.state.sendAsset, this.state.sendAmount, this.state.sendTo, this.state.sendStealth)
                                        }}>Send</Button>
                                    </Col>
                                </FormGroup>
                            </Form>
                        </Col>
                    </Row>
                </Container>);
                break;
            case 'issue':
                mainPage = (<Container>
                    <Row>
                        <Col>
                            <h2>Issue a new asset</h2>
                            <Form>
                                <FormGroup row>
                                    <Label for="ticker" sm={3}>Ticker:</Label>
                                    <Col sm={9}>
                                        <Input type="text" name="ticker" id="ticker" placeholder="Enter ticker symbol" value={this.state.issueTicker} onChange={e => this.setState({ issueTicker: e.target.value.toUpperCase() })} />
                                    </Col>
                                </FormGroup>
                                <FormGroup row>
                                    <Label for="amount" sm={3}>Decimals:</Label>
                                    <Col sm={9}>
                                        <Input type="number" name="decimals" id="decimals" placeholder="Enter decimals of the smallest fraction" value={this.state.issueDecimals} onChange={e => this.setState({ issueDecimals: e.target.value })} />
                                    </Col>
                                </FormGroup>
                                <FormGroup row>
                                    <Label for="supply" sm={3}>Total Supply:</Label>
                                    <Col sm={9}>
                                        <InputGroup>
                                            <Input type="text" name="supply" id="supply" placeholder="Enter total supply to mint (in whole coins)" value={this.state.issueSupply} onChange={e => this.setState({ issueSupply: e.target.value })} />
                                            <InputGroupAddon addonType="append">{this.state.issueTicker}</InputGroupAddon>
                                        </InputGroup>
                                    </Col>
                                </FormGroup>
                                <FormGroup row>
                                    <Col>
                                        <Button onClick={(e) => {
                                            this.issueAsset(this.state.issueTicker, this.state.issueDecimals, this.state.issueSupply)
                                        }}>Issue {this.state.issueTicker}</Button>
                                    </Col>
                                </FormGroup>
                            </Form>
                        </Col>
                    </Row>
                </Container>);
                break;
        }

        var networkBadge = "";
        if(this.state.network !== "MAINNET") {
            networkBadge = (<Badge color="danger">{this.state.network}</Badge>);
        }
        
        var statusIcon = (<div class="nav-link"> | <MdRefresh /></div>)
        if(this.state.synced) {
            statusIcon = (<div class="nav-link"> | <MdCheckCircle/></div>)
        }

        

        return (
            <div>
                <Navbar color="inverse" light expand="md">
                    <NavbarBrand href="#"  onClick={(e) => { this.setState({page:'home'}) }}><img alt="Logo" src="logo.svg" /> Extras {networkBadge}</NavbarBrand>
                    <NavbarToggler onClick={this.toggle} />
                    <Collapse isOpen={this.state.isOpen} navbar>
                        <Nav className="ml-auto" navbar>
                            <NavItem>
                                <NavLink href="#" onClick={(e) => { this.setState({page:'home'}) }}>Home</NavLink>
                            </NavItem>
                            <NavItem>
                                <NavLink href="#" onClick={(e) => { this.setState({page:'receive'}) }}>Receive</NavLink>
                            </NavItem>
                            <NavItem>
                                <NavLink href="#" onClick={(e) => { this.setState({page:'send'}) }}>Send</NavLink>
                            </NavItem>
                            <NavItem>
                                <NavLink href="#" onClick={(e) => { this.setState({page:'issue'}) }}>Issue</NavLink>
                            </NavItem>
                            <NavItem>
                                <div class="nav-link">
                                    | Total Balance: <b>{ coins.toString() }</b>.<small>{ fractions.toString() }</small>&nbsp;<b>VTC</b>
                                </div>
                            </NavItem>
                            <NavItem>
                                {statusIcon}
                            </NavItem>
                        </Nav>
                    </Collapse>
                </Navbar>
                <Jumbotron>
                    <Container>
                        <Row>
                            <Col>
                                {mainPage}
                            </Col>
                        </Row>
                    </Container>
                </Jumbotron>
            </div>
        );
    }
}

export default App;