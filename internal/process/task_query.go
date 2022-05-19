package process

import (
	"context"
	"git.internal.yunify.com/qxp/misc/error2"
	"git.internal.yunify.com/qxp/misc/logger"
	"git.internal.yunify.com/qxp/process/internal/component"
	"git.internal.yunify.com/qxp/process/internal/models"
	"git.internal.yunify.com/qxp/process/pkg/client"
	"git.internal.yunify.com/qxp/process/pkg/code"
	"git.internal.yunify.com/qxp/process/pkg/page"
	"gorm.io/gorm"
	"strings"
)

type nodeFunc func(map[string]component.INode)

func nodePack(nfs ...nodeFunc) {
	for _, nf := range nfs {
		nf(component.NodeHandlers)
	}
}

func withStart(node *component.Node) nodeFunc {
	sNode := component.StartNode{
		Node: node,
	}
	return func(m map[string]component.INode) {
		m[start] = &sNode
	}
}

func withNonModel(task *task, node *component.Node) nodeFunc {
	sNode := component.NonModelNode{
		Node:            node,
		Identity:        task.identity,
		TaskRepo:        task.taskRepo,
		HistoryTaskRepo: task.historyTaskRepo,
		InstanceRepo:    task.instantRepo,
	}
	return func(m map[string]component.INode) {
		m[nonModel] = &sNode
	}
}

func withFree(task *task, node *component.Node) nodeFunc {
	fNode := component.FreeNode{
		Node:             node,
		VariablesRepo:    task.variablesRepo,
		TaskRepo:         task.taskRepo,
		HistoryTaskRepo:  task.historyTaskRepo,
		IdentityLinkRepo: task.identityLinkRepo,
	}
	return func(m map[string]component.INode) {
		m[free] = &fNode
	}
}

func withUser(task *task, node *component.Node) component.UserNode {
	uNode := component.UserNode{
		Node:             node,
		VariablesRepo:    task.variablesRepo,
		TaskRepo:         task.taskRepo,
		HistoryTaskRepo:  task.historyTaskRepo,
		IdentityLinkRepo: task.identityLinkRepo,
		InstanceRepo:     task.instantRepo,
	}
	return uNode
}

func withMultiUser(task *task, node *component.Node) nodeFunc {
	ident, _ := client.NewIdentity(task.conf)
	uNode := withUser(task, node)
	mNode := component.MultiUserNode{
		Conf:            task.conf,
		Node:            node,
		UserNode:        uNode,
		Identity:        ident,
		HistoryTaskRepo: task.historyTaskRepo,
	}
	return func(m map[string]component.INode) {
		m[user] = &uNode
		m[multiUser] = &mNode
	}
}

func withParallelGateway(task *task, node *component.Node) nodeFunc {
	c, _ := client.NewCondition(task.conf)
	pNode := component.ParallelGatewayNode{
		Node:             node,
		Condition:        c,
		TaskRepo:         task.taskRepo,
		VariablesRepo:    task.variablesRepo,
		IdentityLinkRepo: task.identityLinkRepo,
	}
	return func(m map[string]component.INode) {
		m[parallelGateway] = &pNode
	}
}

func withInclusiveGateway(task *task, node *component.Node) nodeFunc {
	iNode := component.InclusiveGatewayNode{
		Node:     node,
		TaskRepo: task.taskRepo,
	}
	return func(m map[string]component.INode) {
		m[inclusiveGateway] = &iNode
	}
}

func withService(task *task, node *component.Node) nodeFunc {
	sNode := component.ServiceNode{
		Node:     node,
		TaskRepo: task.taskRepo,
	}
	return func(m map[string]component.INode) {
		m[service] = &sNode
	}
}

func withEnd(task *task, node *component.Node) nodeFunc {
	eNode := component.EndNode{
		Node:            node,
		TaskRepo:        task.taskRepo,
		HistoryTaskRepo: task.historyTaskRepo,
		InstanceRepo:    task.instantRepo,
	}
	return func(m map[string]component.INode) {
		m[end] = &eNode
	}
}

func (t *task) AgencyTask(ctx context.Context, req *QueryTaskReq) (*page.Page, error) {
	cd := &models.QueryTaskCondition{
		ProcessID:  req.ProcessID,
		InstanceID: req.InstanceID,
		NodeDefKey: req.NodeDefKey,
		Name:       req.Name,
		Des:        strings.Join(req.Des, "|"),
		UserID:     req.Assignee,
		DueTime:    req.DueTime,
		Order:      req.Order,
		TaskID:     req.TaskID,
	}
	// query groupID
	if req.Assignee != "" {
		depID, err := t.QueryGroupIDByUserID(ctx, req.Assignee)
		if err != nil {
			return nil, err
		}
		cd.GroupID = depID
	}
	list, total := t.taskRepo.FindPageByCondition(t.db, req.Page, req.Limit, cd)
	pages := page.NewPage(req.Page, req.Limit, total)
	pages.Data = list
	pages.TotalCount = total
	return pages, nil
}

func (t *task) AgencyTaskTotal(ctx context.Context, req *QueryTaskReq) (*TotalTaskResp, error) {
	cd := &models.QueryTaskCondition{
		ProcessID:  req.ProcessID,
		InstanceID: req.InstanceID,
		NodeDefKey: req.NodeDefKey,
		Name:       req.Name,
		Des:        strings.Join(req.Des, "|"),
		UserID:     req.Assignee,
		DueTime:    req.DueTime,
		Order:      req.Order,
		TaskID:     req.TaskID,
	}
	resp := &TotalTaskResp{}
	// query groupID
	if req.Assignee != "" {
		depID, err := t.QueryGroupIDByUserID(ctx, req.Assignee)
		if err != nil {
			return resp, err
		}
		cd.GroupID = depID
	}
	_, total := t.taskRepo.FindPageByCondition(t.db, 0, 1, cd)
	resp.Total = total
	return resp, nil
}

func (t *task) WholeTask(ctx context.Context, req *QueryTaskReq) (*page.Page, error) {
	cd := &models.QueryTaskCondition{
		ProcessID:  req.ProcessID,
		InstanceID: req.InstanceID,
		NodeDefKey: req.NodeDefKey,
		Name:       req.Name,
		Des:        strings.Join(req.Des, "|"),
		UserID:     req.Assignee,
		DueTime:    req.DueTime,
		Order:      req.Order,
		Status:     req.Status,
	}
	// query groupID
	if req.Assignee != "" {
		depID, err := t.QueryGroupIDByUserID(ctx, req.Assignee)
		if err != nil {
			return nil, err
		}
		cd.GroupID = depID
	}
	list, total := t.taskRepo.FindAllPageByCondition(t.db, req.Page, req.Limit, cd)
	pages := page.NewPage(req.Page, req.Limit, total)
	pages.Data = list
	return pages, nil
}

func (t *task) QueryGroupIDByUserID(ctx context.Context, userID string) (string, error) {
	dep, err := t.identity.FindGroupsByUserID(ctx, userID)
	if err != nil {
		return "", err
	}
	if dep == nil {
		return "", error2.NewError(code.InvalidParams)
	}
	return dep.ID, nil
}

func (t *task) InstanceDoneTasks(ctx context.Context, req *QueryTaskReq) (*page.Page, error) {
	cd := &models.QueryTaskCondition{
		InstanceID: req.InstanceID,
		NodeDefKey: req.NodeDefKey,
		Name:       req.Name,
		Assignee:   req.Assignee,
		Des:        strings.Join(req.Des, "|"),
		Order:      req.Order,
	}
	list, total := t.historyTaskRepo.FindByInstanceID(t.db, req.Page, req.Limit, cd)
	pages := page.NewPage(req.Page, req.Limit, total)
	pages.Data = list
	return pages, nil
}

func (t *task) FindTaskPreNode(ctx context.Context, req *TaskPreNodeReq) (*TaskPreNodeResp, error) {
	tk, err := t.nodeInstanceRepo.FindByTaskID(t.db, req.TaskID)
	if err != nil {
		return nil, err
	}
	ns := make([]*nodeInfo, 0)
	if err := t.findNode(tk.PID, &ns); err != nil {
		return nil, err
	}
	return &TaskPreNodeResp{
		Nodes: ns,
	}, nil
}

func (t *task) findNode(nodeInstanceID string, ns *[]*nodeInfo) error {
	ni, err := t.nodeInstanceRepo.FindByID(t.db, nodeInstanceID)
	if err != nil {
		return err
	}
	if ni != nil {
		// 判断是否是加签的节点,加签节点属于实例节点不属于流程节点，加签的节点不能回退
		nd, err := t.nodeRepo.FindByProcessID(t.db, ni.ProcID, ni.NodeDefKey)
		if err != nil {
			return err
		}
		if len(nd) > 0 {
			nInfo := &nodeInfo{
				Name:       ni.NodeName,
				NodeDefKey: ni.NodeDefKey,
				NodeType:   ni.NodeType,
			}
			*ns = append(*ns, nInfo)
		}
		if ni.NodeType != start && ni.NodeType != parallelGateway && ni.NodeType != inclusiveGateway {
			err := t.findNode(ni.PID, ns)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *task) FindGateWayExecutions(ctx context.Context, req *GateWayExecutionReq) (*GateWayExecutionResp, error) {
	task, err := t.taskRepo.FindByID(t.db, req.TaskID)
	if err != nil {
		return nil, err
	}
	node, err := t.nodeRepo.FindByID(t.db, task.NodeID)
	if err != nil {
		return nil, err
	}
	ex, err := t.executionRepo.FindByID(t.db, task.ExecutionID)
	if err != nil {
		return nil, err
	}
	currentEx := ex.ID
	exID := ex.PID
	if node.NodeType == multiUser && exID != "" {
		ex1, err := t.executionRepo.FindByID(t.db, ex.PID)
		if err != nil {
			return nil, err
		}
		currentEx = ex1.ID
		exID = ex1.PID
	}
	if exID == "" {
		logger.Logger.Error("查询网关分支，pid不存在", logger.STDRequestID(ctx))
		return nil, error2.NewError(code.NoBranchExecution)
	}
	exs, err := t.executionRepo.FindAllByPID(t.db, task.ProcInstanceID, exID)
	if err != nil {
		return nil, err
	}
	exids := make([]string, 0)
	for _, e := range exs {
		if e.ID == currentEx {
			continue
		}
		exids = append(exids, e.ID)
	}
	return &GateWayExecutionResp{
		Executions:         exids,
		CurrentExecutionID: currentEx,
	}, nil
}

func (t *task) FindParentExecutions(ctx context.Context, req *ParentExecutionReq) (*ParentExecutionResp, error) {
	task, err := t.taskRepo.FindByID(t.db, req.TaskID)
	if err != nil {
		return nil, err
	}
	node, err := t.nodeRepo.FindByID(t.db, task.NodeID)
	if err != nil {
		return nil, err
	}
	ex, err := t.executionRepo.FindByID(t.db, task.ExecutionID)
	if err != nil {
		return nil, err
	}
	currentEx := ex.ID
	if node.NodeType == multiUser && ex.PID != "" {
		ex1, err := t.executionRepo.FindByID(t.db, ex.PID)
		if err != nil {
			return nil, err
		}
		currentEx = ex1.ID
	}
	ns := make([]*models.Node, 0)
	if err = t.findInclusiveGateway(node.ProcID, node.ID, req.DefKey, &ns); err != nil {
		return nil, err
	}
	for i := 0; i < len(ns); i++ {
		exp, err := t.executionRepo.FindByID(t.db, currentEx)
		if err != nil {
			return nil, err
		}
		currentEx = exp.PID
	}
	return &ParentExecutionResp{
		ExecutionID: currentEx,
	}, nil
}

func (t *task) findInclusiveGateway(processID, nodeID, defKey string, ns *[]*models.Node) error {
	nl, err := t.nodeLinkRepo.FindByNodeID(t.db, processID, nodeID)
	if err != nil {
		return err
	}
	if len(nl) == 0 {
		logger.Logger.Info("该节点没有下个节点")
		return error2.NewError(code.InvalidParams)
	}
	node, err := t.nodeRepo.FindByDefKey(t.db, processID, nl[0].NextNodeDefKey)
	if err != nil {
		return err
	}
	if node.DefKey == defKey {
		return nil
	}
	if node.NodeType == inclusiveGateway {
		*ns = append(*ns, node)
	}
	err = t.findInclusiveGateway(processID, node.ID, defKey, ns)
	return err
}

func (t *task) findActiveExecution(tx *gorm.DB, instanceID, executionID string, res *[]*models.Execution) error {
	exps, err := t.executionRepo.FindByPID(tx, instanceID, executionID, 1)
	if err != nil {
		return err
	}
	if len(exps) == 0 {
		// 还需要递归查找同级分支的子分支是否有未完成的
		childEx := make([]*models.Execution, 0)
		expsa, err := t.executionRepo.FindAllByPID(tx, instanceID, executionID)
		if err != nil {
			return err
		}
		for i := 0; i < len(expsa); i++ {
			err = t.findActiveExecution(tx, expsa[i].ProcInstanceID, expsa[i].ID, &childEx)
			if err != nil {
				return err
			}
		}
		exps = append(exps, childEx...)
	}
	*res = append(*res, exps...)
	return nil
}

func (t *task) findChildExecution(instanceID, executionID string, res *[]*models.Execution) error {
	exs, err := t.executionRepo.FindByPID(t.db, instanceID, executionID, 1)
	if err != nil {
		return err
	}
	if len(exs) > 0 {
		for _, ex := range exs {
			*res = append(*res, ex)
			err := t.findChildExecution(ex.ProcInstanceID, ex.ID, res)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *task) InstanceTasks(ctx context.Context, req *InstanceTaskReq) ([]*models.Task, error) {
	res, err := t.taskRepo.FindByInstanceID(t.db, req.InstanceID)
	return res, err
}
